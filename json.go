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
	InitialPacketJSON:            func() interface{} { return new(InitialPacket) },
	RetryPacketJSON:              func() interface{} { return new(RetryPacket) },
	StatelessResetPacketJSON:     func() interface{} { return new(StatelessResetPacket) },
	VersionNegotiationPacketJSON: func() interface{} { return new(VersionNegotiationPacket) },
	HandshakePacketJSON:          func() interface{} { return new(HandshakePacket) },
	ProtectedPacketJSON:          func() interface{} { return new(ProtectedPacket) },
	ZeroRTTProtectedPacketJSON:   func() interface{} { return new(ZeroRTTProtectedPacket) },

	ShortHeaderJSON: func() interface{} { return new(ShortHeader) },
	LongHeaderJSON:  func() interface{} { return new(LongHeader) },

	PaddingFrameJSON:           func() interface{} { return new(PaddingFrame) },
	PingFrameJSON:              func() interface{} { return new(PingFrame) },
	AckFrameJSON:               func() interface{} { return new(AckFrame) },
	AckECNFrameJSON:            func() interface{} { return new(AckECNFrame) },
	ResetStreamJSON:            func() interface{} { return new(ResetStream) },
	StopSendingFrameJSON:       func() interface{} { return new(StopSendingFrame) },
	CryptoFrameJSON:            func() interface{} { return new(CryptoFrame) },
	NewTokenFrameJSON:          func() interface{} { return new(NewTokenFrame) },
	StreamFrameJSON:            func() interface{} { return new(StreamFrame) },
	MaxDataFrameJSON:           func() interface{} { return new(MaxDataFrame) },
	MaxStreamsFrameJSON:        func() interface{} { return new(MaxStreamsFrame) },
	MaxStreamDataFrameJSON:     func() interface{} { return new(MaxStreamDataFrame) },
	DataBlockedFrameJSON:       func() interface{} { return new(DataBlockedFrame) },
	StreamDataBlockedFrameJSON: func() interface{} { return new(StreamDataBlockedFrame) },
	StreamsBlockedFrameJSON:    func() interface{} { return new(StreamsBlockedFrame) },
	NewConnectionIdFrameJSON:   func() interface{} { return new(NewConnectionIdFrame) },
	RetireConnectionIdJSON:     func() interface{} { return new(RetireConnectionId) },
	PathChallengeJSON:          func() interface{} { return new(PathChallenge) },
	PathResponseJSON:           func() interface{} { return new(PathResponse) },
	ConnectionCloseFrameJSON:   func() interface{} { return new(ConnectionCloseFrame) },
	ApplicationCloseFrameJSON:  func() interface{} { return new(ApplicationCloseFrame) },
	HandshakeDoneFrameJSON:     func() interface{} { return new(HandshakeDoneFrame) },
}

type Envelope struct {
	Type    JSONType
	Message interface{}
}

func (p *Envelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Envelope) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}

	msg := JSONTypeHandlers[p.Type]()
	if err := ms.Decode(p.Message, msg); err != nil {
		return err
	}

	p.Message = msg
	return nil
}
