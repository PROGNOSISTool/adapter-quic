package adapter

import (
	"encoding/json"
	"fmt"
	ms "github.com/tiferrei/golang-set"
	qt "github.com/tiferrei/quic-tracker"
	"log"
	"sort"
	"strings"
)

type ConcreteSymbol struct {
	qt.Packet
}

func (cs *ConcreteSymbol) UnmarshalJSON(data []byte) error {
	envelope := qt.Envelope{}
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		log.Printf("Failed to unmarshal ConcreteSymbol: %v", err)
		return err
	}

	*cs = ConcreteSymbol{envelope.Message.(qt.Packet)}
	return nil
}


func NewConcreteSymbol(packet qt.Packet) ConcreteSymbol {
	cs := ConcreteSymbol{packet}
	return cs
}

func (cs *ConcreteSymbol) String() string {
	ba, err := json.Marshal(cs)
	if err != nil {
		fmt.Printf("Failed to Marshal ConcreteSymbol: %v", err.Error())
	}

	return string(ba)
}

type ConcreteSet struct {
	ms.Set // type: ConcreteSymbol
}

func (cs *ConcreteSet) UnmarshalJSON(data []byte) error {
	type jsonSet []ConcreteSymbol
	var internal jsonSet
	err := json.Unmarshal(data, &internal)
	if err != nil {
		fmt.Printf("Failed to Unmarshal ConcreteSet: %v\n", err.Error())
		return err
	}

	interfaceArray := make([]interface{}, len(internal))
	for _, value := range internal {
		interfaceArray = append(interfaceArray, value)
	}

	cs = &ConcreteSet{ms.NewSetFromSlice(interfaceArray)}

	return nil
}

func NewConcreteSet() *ConcreteSet {
	cs := ConcreteSet{ms.NewSet()}
	return &cs
}

func (cs *ConcreteSet) String() string {
	if cs.Cardinality() == 0 {
		return "{}"
	}

	setSlice := cs.ToSlice()
	stringSlice := []string{}
	for _, setElement := range setSlice {
		symbol := setElement.(ConcreteSymbol)
		stringSlice = append(stringSlice, (&symbol).String())
	}
	sort.Strings(stringSlice)

	return fmt.Sprintf("{%v}", strings.Join(stringSlice, ","))
}

type ConcreteOrderedPair struct {
	ConcreteInputs  []*ConcreteSymbol
	ConcreteOutputs []ConcreteSet
}

func (ct *ConcreteOrderedPair) Input() *[]*ConcreteSymbol {
	return &ct.ConcreteInputs
}

func (ct *ConcreteOrderedPair) Output() *[]ConcreteSet {
	return &ct.ConcreteOutputs
}

func (ct *ConcreteOrderedPair) SetInput(concreteSymbols []*ConcreteSymbol) {
	(*ct).ConcreteInputs = concreteSymbols
}

func (ct *ConcreteOrderedPair) SetOutput(concreteSets []ConcreteSet) {
	(*ct).ConcreteOutputs = concreteSets
}

func (ct *ConcreteOrderedPair) String() string {
	ciStringSlice := []string{}
	for _, value := range ct.ConcreteInputs {
		if value != nil {
			ciStringSlice = append(ciStringSlice, value.String())
		} else {
			ciStringSlice = append(ciStringSlice, "NIL")
		}

	}
	ciString := fmt.Sprintf("[%v]", strings.Join(ciStringSlice, ","))

	coStringSlice := []string{}
	for _, value := range ct.ConcreteOutputs {
		coStringSlice = append(coStringSlice, value.String())
	}
	coString := fmt.Sprintf("[%v]", strings.Join(coStringSlice, ","))
	return fmt.Sprintf("(%v,%v)", ciString, coString)
}
