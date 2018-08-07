/*
    Maxime Piraux's master's thesis
    Copyright (C) 2017-2018  Maxime Piraux

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License version 3
	as published by the Free Software Foundation.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package scenarii

import (
	m "github.com/mpiraux/master-thesis"
	"fmt"
)

const (
	SSRS_TLSHandshakeFailed               = 1
	SSRS_DidNotCloseTheConnection         = 2
	SSRS_CloseTheConnectionWithWrongError = 3
	SSRS_MaxStreamUniTooLow				  = 4
	SSRS_UnknownError					  = 5
)

type StopSendingOnReceiveStreamScenario struct {
	AbstractScenario
}

func NewStopSendingOnReceiveStreamScenario() *StopSendingOnReceiveStreamScenario {
	return &StopSendingOnReceiveStreamScenario{AbstractScenario{"stop_sending_frame_on_receive_stream", 1, false}}
}

func (s *StopSendingOnReceiveStreamScenario) Run(conn *m.Connection, trace *m.Trace, preferredUrl string, debug bool) {
	var p m.Packet
	if p, err := CompleteHandshake(conn); err != nil {
		trace.MarkError(SSRS_TLSHandshakeFailed, err.Error(), p)
		return
	}

	if conn.TLSTPHandler.ReceivedParameters.MaxUniStreams < 1 {
		trace.MarkError(SSRS_MaxStreamUniTooLow, "", p)
		trace.Results["expected_max_stream_uni"] = ">= 1"
		trace.Results["received_max_stream_uni"] = conn.TLSTPHandler.ReceivedParameters.MaxUniStreams
		return
	}

	conn.SendHTTPGETRequest(preferredUrl, 2)
	conn.FrameQueue.Submit(&m.StopSendingFrame{StreamId: 2, ErrorCode: 42})

	incPackets := make(chan interface{}, 1000)
	conn.IncomingPackets.Register(incPackets)

	trace.ErrorCode = SSRS_DidNotCloseTheConnection
	for {
		p := (<-incPackets).(m.Packet)

		if framer, ok := p.(m.Framer); ok && framer.Contains(m.ConnectionCloseType){
			cc := framer.GetFirst(m.ConnectionCloseType).(*m.ConnectionCloseFrame)
			if cc.ErrorCode != m.ERR_PROTOCOL_VIOLATION {
				trace.MarkError(SSRS_CloseTheConnectionWithWrongError, "", p)
				trace.Results["connection_closed_error_code"] = fmt.Sprintf("0x%x", cc.ErrorCode)
				return
			}
			trace.ErrorCode = 0
			return
		}
	}
}
