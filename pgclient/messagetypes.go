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
package pgclient

/*
PgOutputType represents the type of an output message to the server.
*/
type PgOutputType int

//go:generate stringer -type PgOutputType messagetypes.go

// Constants representing output message types.
const (
	Bind            PgOutputType = 'B'
	Close           PgOutputType = 'C'
	DescribeMsg     PgOutputType = 'D'
	Execute         PgOutputType = 'E'
	Flush           PgOutputType = 'H'
	Parse           PgOutputType = 'P'
	Query           PgOutputType = 'Q'
	Sync            PgOutputType = 'S'
	Terminate       PgOutputType = 'X'
	CopyDoneOut     PgOutputType = 'c'
	CopyDataOut     PgOutputType = 'd'
	PasswordMessage PgOutputType = 'p'
)

/*
PgInputType is the one-byte type of a postgres response from the server.
*/
type PgInputType int

//go:generate stringer -type PgInputType messagetypes.go

// Various types of messages that represent one-byte message types.
const (
	ErrorResponse          PgInputType = 'E'
	CommandComplete        PgInputType = 'C'
	DataRow                PgInputType = 'D'
	CopyInResponse         PgInputType = 'G'
	CopyOutResponse        PgInputType = 'H'
	EmptyQueryResponse     PgInputType = 'I'
	BackEndKeyData         PgInputType = 'K'
	NoticeResponse         PgInputType = 'N'
	AuthenticationResponse PgInputType = 'R'
	ParameterStatus        PgInputType = 'S'
	RowDescription         PgInputType = 'T'
	CopyBothResponse       PgInputType = 'W'
	ReadyForQuery          PgInputType = 'Z'
	CopyDone               PgInputType = 'c'
	CopyData               PgInputType = 'd'
	HotStandbyFeedback     PgInputType = 'h'
	SenderKeepalive        PgInputType = 'k'
	NoData                 PgInputType = 'n'
	StandbyStatusUpdate    PgInputType = 'r'
	PortalSuspended        PgInputType = 's'
	ParameterDescription   PgInputType = 't'
	WALData                PgInputType = 'w'
	ParseComplete          PgInputType = '1'
	BindComplete           PgInputType = '2'
	CloseComplete          PgInputType = '3'
)

/*
PgType represents a Postgres type OID.
*/
type PgType int

//go:generate stringer -type PgType messagetypes.go

// Constants for well-known OIDs that we care about
const (
	Bytea       PgType = 17
	Int8        PgType = 20
	Int2        PgType = 21
	Int4        PgType = 23
	OID         PgType = 26
	Float4      PgType = 700
	Float8      PgType = 701
	Timestamp   PgType = 1114
	TimestampTZ PgType = 1184
)
