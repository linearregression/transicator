/*
Copyright 2016 The Transicator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/30x/goscaffold"
	log "github.com/Sirupsen/logrus"
	"github.com/apigee-labs/transicator/common"
	"github.com/apigee-labs/transicator/replication"
	"github.com/apigee-labs/transicator/storage"
	"github.com/julienschmidt/httprouter"
)

const (
	jsonContent  = "application/json"
	protoContent = "application/transicator+protobuf"
	textContent  = "text/plain"

	// internalScope is a scope we'll use to track sequences on delete. It
	// should never show up in data that we get from clients.
	internalScope = "__transicator_internal"

	defaultSelectorColumn = "_change_selector"
)

var selectorColumn = defaultSelectorColumn

type server struct {
	db          storage.DB
	repl        *replication.Replicator
	tracker     *changeTracker
	cleaner     *cleaner
	firstChange common.Sequence
	slotName    string
	dropSlot    int32
	stopChan    chan chan<- bool
}

type errMsg struct {
	Error string `json:"error"`
}

func createChangeServer(mux *http.ServeMux, dbDir, pgURL, pgSlot, urlPrefix string) (*server, error) {
	success := false
	slotName := sanitizeSlotName(pgSlot)

	db, err := storage.Open(dbDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			db.Close()
		}
	}()

	// Retrieve the highest sequence from the DB so that we don't
	// process duplicate updates that might come from PG.
	_, _, firstChange, err := db.Scan(nil, 0, 0, 0, nil)

	repl, err := replication.CreateReplicator(pgURL, slotName)
	if err != nil {
		return nil, err
	}

	success = true

	s := &server{
		db:          db,
		repl:        repl,
		slotName:    slotName,
		firstChange: firstChange,
		tracker:     createTracker(),
		stopChan:    make(chan chan<- bool, 1),
	}

	router := httprouter.New()
	mux.Handle("/", router)

	s.initChangesAPI(urlPrefix, router)
	s.initDiagAPI(urlPrefix, router)

	return s, nil
}
func (s *server) start() {
	s.repl.Start()
	go s.runReplication(s.firstChange)
}

func (s *server) stop() {
	if s.cleaner != nil {
		s.cleaner.stop()
	}
	stopped := make(chan bool, 1)
	s.stopChan <- stopped
	<-stopped
	s.tracker.close()
	s.db.Close()
}

func (s *server) delete() error {
	return s.db.Delete()
}

func (s *server) checkHealth() (goscaffold.HealthStatus, error) {
	// Scan the first and last sequence numbers from the DB
	_, _, _, err := s.db.Scan(nil, 0, 0, 0, nil)
	if err == nil {
		return goscaffold.OK, nil
	}
	// If we get an error reading from LevelDB, things are really bad.
	// Mark ourselves "unhealthy" and we may get restarted
	return goscaffold.Failed, err
}

func getIntParam(q url.Values, key string, dflt int) (int, error) {
	qs := q.Get(key)
	if qs == "" {
		return dflt, nil
	}
	v, err := strconv.ParseInt(qs, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func sendError(resp http.ResponseWriter, req *http.Request, code int, msg string) {
	log.Debugf("sendError: code = %d msg = %s req = %v", code, msg, req)
	ct := goscaffold.SelectMediaType(req, []string{jsonContent, textContent})

	switch ct {
	case jsonContent:
		em := &errMsg{
			Error: msg,
		}
		eb, _ := json.Marshal(em)
		resp.Header().Set("Content-Type", jsonContent)
		resp.WriteHeader(code)
		resp.Write(eb)

	default:
		resp.Header().Set("Content-Type", textContent)
		resp.WriteHeader(code)
		resp.Write([]byte(msg))
	}
}

/*
sanitizeSlotName converts the name from the "slot" API parameter to a name
that will actually work in Postgres. Postgres slot names can only contain
upper and lower case letters, numbers, and underscores. This method
converts dashes to underscores, and removes everything else.
*/
func sanitizeSlotName(name string) string {
	return strings.Map(func(c rune) rune {
		switch {
		case
			c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '_':
			return c
		case c == '-':
			return '_'
		default:
			return -1
		}
	}, name)
}
