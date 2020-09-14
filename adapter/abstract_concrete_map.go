package adapter

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// The key is a stringified AbstractOrderedPair.
// Because in Go slices aren't Comparable and thus can't be map keys. I know, I'm angry too.
type AbstractConcreteMap map[string]ConcreteOrderedPair

func NewAbstractConcreteMap() *AbstractConcreteMap {
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

	dataDecoder := gob.NewDecoder(gobFile)
	err = dataDecoder.Decode(acm)
	if err != nil {
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

	dataEncoder := gob.NewEncoder(dataFile)
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
