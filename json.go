package quictracker

import (
	"encoding/json"
	ms "github.com/mitchellh/mapstructure"
)

//go:generate jsonenums -type=JSONType

type JSONType int

const (
	InitialPacketJSON JSONType = iota
	RetryPacketJSON
	StatelessResetPacketJSON
	VersionNegotiationPacketJSON
	HandshakePacketJSON
	ProtectedPacketJSON
	ZeroRTTProtectedPacketJSON

	ShortHeaderJSON
	LongHeaderJSON

	PaddingFrameJSON
	PingFrameJSON
	AckFrameJSON
	AckECNFrameJSON
	ResetStreamJSON
	StopSendingFrameJSON
	CryptoFrameJSON
	NewTokenFrameJSON
	StreamFrameJSON
	MaxDataFrameJSON
	MaxStreamsFrameJSON
	MaxStreamDataFrameJSON
	DataBlockedFrameJSON
	StreamDataBlockedFrameJSON
	StreamsBlockedFrameJSON
	NewConnectionIdFrameJSON
	RetireConnectionIdJSON
	PathChallengeJSON
	PathResponseJSON
	ConnectionCloseFrameJSON
	ApplicationCloseFrameJSON
	HandshakeDoneFrameJSON
)

var JSONTypeHandlers = map[JSONType]func() interface{} {
	InitialPacketJSON:            func() interface{} { type local InitialPacket; return new(local) },
	RetryPacketJSON:              func() interface{} { type local RetryPacket; return new(local) },
	StatelessResetPacketJSON:     func() interface{} { type local StatelessResetPacket; return new(local) },
	VersionNegotiationPacketJSON: func() interface{} { type local VersionNegotiationPacket; return new(local) },
	HandshakePacketJSON:          func() interface{} { type local HandshakePacket; return new(local) },
	ProtectedPacketJSON:          func() interface{} { type local ProtectedPacket; return new(local) },
	ZeroRTTProtectedPacketJSON:   func() interface{} { type local ZeroRTTProtectedPacket; return new(local) },

	ShortHeaderJSON: func() interface{} { type local ShortHeader; return new(local) },
	LongHeaderJSON:  func() interface{} { type local LongHeader; return new(local) },

	PaddingFrameJSON:           func() interface{} { type local PaddingFrame; return new(local) },
	PingFrameJSON:              func() interface{} { type local PingFrame; return new(local) },
	AckFrameJSON:               func() interface{} { type local AckFrame; return new(local) },
	AckECNFrameJSON:            func() interface{} { type local AckECNFrame; return new(local) },
	ResetStreamJSON:            func() interface{} { type local ResetStream; return new(local) },
	StopSendingFrameJSON:       func() interface{} { type local StopSendingFrame; return new(local) },
	CryptoFrameJSON:            func() interface{} { type local CryptoFrame; return new(local) },
	NewTokenFrameJSON:          func() interface{} { type local NewTokenFrame; return new(local) },
	StreamFrameJSON:            func() interface{} { type local StreamFrame; return new(local) },
	MaxDataFrameJSON:           func() interface{} { type local MaxDataFrame; return new(local) },
	MaxStreamsFrameJSON:        func() interface{} { type local MaxStreamsFrame; return new(local) },
	MaxStreamDataFrameJSON:     func() interface{} { type local MaxStreamDataFrame; return new(local) },
	DataBlockedFrameJSON:       func() interface{} { type local DataBlockedFrame; return new(local) },
	StreamDataBlockedFrameJSON: func() interface{} { type local StreamDataBlockedFrame; return new(local) },
	StreamsBlockedFrameJSON:    func() interface{} { type local StreamsBlockedFrame; return new(local) },
	NewConnectionIdFrameJSON:   func() interface{} { type local NewConnectionIdFrame; return new(local) },
	RetireConnectionIdJSON:     func() interface{} { type local RetireConnectionId; return new(local) },
	PathChallengeJSON:          func() interface{} { type local PathChallenge; return new(local) },
	PathResponseJSON:           func() interface{} { type local PathResponse; return new(local) },
	ConnectionCloseFrameJSON:   func() interface{} { type local ConnectionCloseFrame; return new(local) },
	ApplicationCloseFrameJSON:  func() interface{} { type local ApplicationCloseFrame; return new(local) },
	HandshakeDoneFrameJSON:     func() interface{} { type local HandshakeDoneFrame; return new(local) },
}

type Envelope struct {
	Type    JSONType
	Message interface{}
}

func (p *Envelope) UnmarshalJSON(data []byte) error {
	tempEnvelope := map[string]interface{}{}
	if err := json.Unmarshal(data, &tempEnvelope); err != nil {
		return err
	}

	msg := JSONTypeHandlers[p.Type]()
	if err := ms.Decode(tempEnvelope["Message"], msg); err != nil {
		return err
	}

	p.Message = msg
	return nil
}
