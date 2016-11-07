package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/30x/goscaffold"
	"github.com/apigee-labs/transicator/common"
	"github.com/apigee-labs/transicator/replication"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

const (
	defaultLimit = 100
)

func (s *server) initChangesAPI(prefix string, router *httprouter.Router) {
	router.HandlerFunc("GET", prefix+"/changes", s.handleGetChanges)
}

func (s *server) handleGetChanges(resp http.ResponseWriter, req *http.Request) {
	enc := goscaffold.SelectMediaType(req, []string{jsonContent, protoContent})
	if enc == "" {
		resp.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	q := req.URL.Query()

	limit, err := getIntParam(q, "limit", defaultLimit)
	if err != nil {
		sendError(resp, req, http.StatusBadRequest, "Invalid limit parameter")
		return
	}

	block, err := getIntParam(q, "block", 0)
	if err != nil {
		sendError(resp, req, http.StatusBadRequest, "Invalid block parameter")
		return
	}

	scopes := q["scope"]
	if len(scopes) == 0 {
		// If no scope specified, replace with the empty scope
		scopes = []string{""}
	}

	var sinceSeq common.Sequence
	since := q.Get("since")
	if since == "" {
		sinceSeq = common.Sequence{}
	} else {
		sinceSeq, err = common.ParseSequence(since)
		if err != nil {
			sendError(resp, req, http.StatusBadRequest, fmt.Sprintf("Invalid since value %s", since))
			return
		}
	}

	var snapshotFilter func([]byte) bool
	snapStr := q.Get("snapshot")
	if snapStr != "" {
		var snapshot *replication.Snapshot
		snapshot, err = replication.MakeSnapshot(snapStr)
		if err != nil {
			sendError(resp, req, http.StatusBadRequest, fmt.Sprintf("Invalid snapshot %s", snapStr))
			return
		}
		snapshotFilter = makeSnapshotFilter(snapshot)
	}

	// Need to advance past a single "since" value
	sinceSeq.Index++
	var metadata [][]byte

	log.Debugf("Receiving changes: scopes = %v since = %s limit = %d block = %d",
		scopes, sinceSeq, limit, block)
	entries, metadata, err := s.db.GetMultiEntries(
		scopes, []string{lastSequenceKey}, sinceSeq.LSN,
		sinceSeq.Index, limit, snapshotFilter)
	if err != nil {
		sendError(resp, req, http.StatusInternalServerError, err.Error())
		return
	}
	log.Debugf("Received %d changes", len(entries))

	lastSeq, err := common.ParseSequenceBytes(metadata[0])
	if err != nil {
		sendError(resp, req, http.StatusInternalServerError, err.Error())
		return
	}

	if len(entries) == 0 && block > 0 {
		// Query -- which was consistent at the "snapshot" level -- didn't
		// return anything. Wait until something is put in the database and try again.
		waitSeq := lastSeq
		waitSeq.Index++

		log.Debugf("Blocking at %s for up to %d seconds", waitSeq, block)
		newIndex := s.tracker.timedWait(waitSeq, time.Duration(block)*time.Second, scopes)
		if newIndex.Compare(sinceSeq) > 0 {
			entries, metadata, err = s.db.GetMultiEntries(
				scopes, []string{lastSequenceKey}, sinceSeq.LSN,
				sinceSeq.Index, limit, snapshotFilter)
			if err != nil {
				sendError(resp, req, http.StatusInternalServerError, err.Error())
				return
			}
		}
		log.Debugf("Received %d changes after blocking", len(entries))

		lastSeq, err = common.ParseSequenceBytes(metadata[0])
		if err != nil {
			sendError(resp, req, http.StatusInternalServerError, err.Error())
			return
		}
	}

	changeList := common.ChangeList{
		LastSequence: lastSeq.String(),
	}

	for _, e := range entries {
		change, err := decodeChangeProto(e)
		if err != nil {
			sendError(resp, req, http.StatusInternalServerError,
				fmt.Sprintf("Invalid data in database: %s", err))
		}
		// Database doesn't have value of "Sequence" in it
		change.Sequence = change.GetSequence().String()
		changeList.Changes = append(changeList.Changes, *change)
	}

	switch enc {
	case jsonContent:
		resp.Header().Set("Content-Type", jsonContent)
		resp.Write(changeList.Marshal())
	case protoContent:
		resp.Header().Set("Content-Type", protoContent)
		resp.Write(changeList.MarshalProto())
	default:
		panic("Got to an unsupported media type")
	}
}

func makeSnapshotFilter(ss *replication.Snapshot) func([]byte) bool {
	return func(buf []byte) bool {
		txid, err := decodeChangeTXID(buf)
		if err == nil {
			return !ss.Contains(txid)
		}
		return false
	}
}
