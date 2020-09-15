package adapter

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil && err != io.EOF {
		fmt.Printf("Failed to open JSON file: %v\n", err.Error())
		return acm
	}

	err = json.Unmarshal(jsonFile, acm)
	if err != nil {
		fmt.Printf("Failed to decode JSON file: %v\n", err.Error())
		return acm
	}

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

func (acm *AbstractConcreteMap) AddOPs(abstractOrderedPair AbstractOrderedPair, concreteOrderedPair ConcreteOrderedPair) {
	(*acm)[abstractOrderedPair.String()] = concreteOrderedPair
}

func (acm *AbstractConcreteMap) AddIOs(abstractInputs []AbstractSymbol, abstractOutputs []AbstractSet, concreteInputs []*ConcreteSymbol, concreteOutputs []ConcreteSet) {
	abstractOP := AbstractOrderedPair{AbstractInputs: abstractInputs, AbstractOutputs: abstractOutputs}
	concreteOP := ConcreteOrderedPair{ConcreteInputs: concreteInputs, ConcreteOutputs: concreteOutputs}
	acm.AddOPs(abstractOP, concreteOP)
}
