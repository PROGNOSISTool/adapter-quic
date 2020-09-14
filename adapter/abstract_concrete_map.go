package adapter

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	mapset "github.com/tiferrei/golang-set"
	qt "github.com/tiferrei/quic-tracker"
	"io"
	"os"
	"strings"
)

// The key is a stringified AbstractOrderedPair.
// Because in Go slices aren't Comparable and thus can't be map keys. I know, I'm angry too.
type AbstractConcreteMap map[string]ConcreteOrderedPair

func NewAbstractConcreteMap() *AbstractConcreteMap {
	gob.Register(AbstractConcreteMap{})
	acm := AbstractConcreteMap{}
	return &acm
}

func ReadAbstractConcreteMap(filename string) *AbstractConcreteMap {
	acm := NewAbstractConcreteMap()

	fmt.Printf("Reading Oracle table...\n")
	gobFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Failed to open GOB file: %v\n", err.Error())
		return acm
	}

	gobRegister()
	dataDecoder := gob.NewDecoder(gobFile)
	err = dataDecoder.Decode(acm)
	if err != nil && err != io.EOF {
		fmt.Printf("Failed to decode GOB file: %v\n", err.Error())
		return acm
	}

	gobFile.Close()
	return acm
}

func (acm *AbstractConcreteMap) String() string {
	var sb strings.Builder
	for key, value := range *acm {
		sb.WriteString(fmt.Sprintf("%v->%v\n", key, value.String()))
	}
	return sb.String()
}

func (acm *AbstractConcreteMap) JSON() string {
	ba, err := json.Marshal(acm)
	if err != nil {
		fmt.Printf("Failed to Marshal AbstractConcreteMap: %v\n", err.Error())
	}
	return string(ba)
}

func (acm *AbstractConcreteMap) SaveToDisk(filename string) error {
	dataFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create GOB file: %v\n", err.Error())
		return err
	}

	gobRegister()
	dataEncoder := gob.NewEncoder(dataFile)
	for key, _ := range *acm {
		fmt.Printf("[DEBUG] %v\n", key)
	}

	err = dataEncoder.Encode(*acm)
	if err != nil {
		fmt.Printf("Failed to encode to GOB file: %v\n", err.Error())
		dataFile.Close()
		return err
	}

	dataFile.Close()
	return nil
}

func (acm *AbstractConcreteMap) AddOPs(abstractOrderedPair AbstractOrderedPair, concreteOrderedPair ConcreteOrderedPair) {
	(*acm)[abstractOrderedPair.String()] = concreteOrderedPair
}

func (acm *AbstractConcreteMap) AddIOs(abstractInputs []AbstractSymbol, abstractOutputs []AbstractSet, concreteInputs []*ConcreteSymbol, concreteOutputs []ConcreteSet) {
	abstractOP := AbstractOrderedPair{AbstractInputs: abstractInputs, AbstractOutputs: abstractOutputs}
	concreteOP := ConcreteOrderedPair{ConcreteInputs: concreteInputs, ConcreteOutputs: concreteOutputs}
	acm.AddOPs(abstractOP, concreteOP)
}

func gobRegister() {
	// Symbols
	gob.Register(ConcreteSet{})
	gob.Register(mapset.NewSet())
	gob.Register(ConcreteSymbol{})

	// Packets
	gob.Register(qt.InitialPacket{})
	gob.Register(qt.RetryPacket{})
	gob.Register(qt.StatelessResetPacket{})
	gob.Register(qt.VersionNegotiationPacket{})
	gob.Register(qt.HandshakePacket{})
	gob.Register(qt.ProtectedPacket{})
	gob.Register(qt.ZeroRTTProtectedPacket{})

	// Headers
	gob.Register(new(qt.ShortHeader))
	gob.Register(new(qt.LongHeader))

	// Streams
	gob.Register(new(qt.PaddingFrame))
	gob.Register(new(qt.PingFrame))
	gob.Register(new(qt.AckFrame))
	gob.Register(new(qt.AckECNFrame))
	gob.Register(new(qt.ResetStream))
	gob.Register(new(qt.StopSendingFrame))
	gob.Register(new(qt.CryptoFrame))
	gob.Register(new(qt.NewTokenFrame))
	gob.Register(new(qt.StreamFrame))
	gob.Register(new(qt.MaxDataFrame))
	gob.Register(new(qt.MaxStreamDataFrame))
	gob.Register(new(qt.MaxStreamsFrame))
	gob.Register(new(qt.DataBlockedFrame))
	gob.Register(new(qt.StreamDataBlockedFrame))
	gob.Register(new(qt.StreamsBlockedFrame))
	gob.Register(new(qt.NewConnectionIdFrame))
	gob.Register(new(qt.RetireConnectionId))
	gob.Register(new(qt.PathChallenge))
	gob.Register(new(qt.PathResponse))
	gob.Register(new(qt.ConnectionCloseFrame))
	gob.Register(new(qt.ApplicationCloseFrame))
	gob.Register(new(qt.HandshakeDoneFrame))
}
