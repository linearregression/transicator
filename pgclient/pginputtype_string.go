// Code generated by "stringer -type PgInputType ."; DO NOT EDIT

package pgclient

import "fmt"

const _PgInputType_name = "ParseCompleteBindCompleteCloseCompleteCommandCompleteDataRowErrorResponseCopyInResponseCopyOutResponseEmptyQueryResponseBackEndKeyDataNoticeResponseAuthenticationResponseParameterStatusRowDescriptionCopyBothResponseReadyForQueryCopyDoneCopyDataHotStandbyFeedbackSenderKeepaliveNoDataStandbyStatusUpdateParameterDescriptionWALData"

var _PgInputType_map = map[PgInputType]string{
	49:  _PgInputType_name[0:13],
	50:  _PgInputType_name[13:25],
	51:  _PgInputType_name[25:38],
	67:  _PgInputType_name[38:53],
	68:  _PgInputType_name[53:60],
	69:  _PgInputType_name[60:73],
	71:  _PgInputType_name[73:87],
	72:  _PgInputType_name[87:102],
	73:  _PgInputType_name[102:120],
	75:  _PgInputType_name[120:134],
	78:  _PgInputType_name[134:148],
	82:  _PgInputType_name[148:170],
	83:  _PgInputType_name[170:185],
	84:  _PgInputType_name[185:199],
	87:  _PgInputType_name[199:215],
	90:  _PgInputType_name[215:228],
	99:  _PgInputType_name[228:236],
	100: _PgInputType_name[236:244],
	104: _PgInputType_name[244:262],
	107: _PgInputType_name[262:277],
	110: _PgInputType_name[277:283],
	114: _PgInputType_name[283:302],
	116: _PgInputType_name[302:322],
	119: _PgInputType_name[322:329],
}

func (i PgInputType) String() string {
	if str, ok := _PgInputType_map[i]; ok {
		return str
	}
	return fmt.Sprintf("PgInputType(%d)", i)
}
