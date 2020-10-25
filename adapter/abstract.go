package adapter

import (
	"fmt"
	mapset "github.com/tiferrei/golang-set"
	qt "github.com/tiferrei/quic-tracker"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var packetTypeToString = map[qt.PacketType]string {
	qt.VersionNegotiation: "VERNEG",
	qt.Initial: "INITIAL",
	qt.Retry: "RETRY",
	qt.Handshake: "HANDSHAKE",
	qt.ZeroRTTProtected: "ZERO",
	qt.ShortHeaderPacket: "SHORT",
	qt.StatelessReset : "RESET",
}

var stringToPacketType = map[string]qt.PacketType {
	"VERNEG": qt.VersionNegotiation,
	"INITIAL": qt.Initial,
	"RETRY": qt.Retry,
	"HANDSHAKE": qt.Handshake,
	"ZERO": qt.ZeroRTTProtected,
	"SHORT": qt.ShortHeaderPacket,
	"RESET": qt.StatelessReset,
}

type HeaderOptions struct {
	PacketNumber *qt.PacketNumber
	QUICVersion *uint32
}

func (ho *HeaderOptions) String() string {
	version := "?"
	if ho.QUICVersion != nil {
		version = fmt.Sprintf("%#x", qt.Uint32ToBEBytes(*ho.QUICVersion))
	}
	packetNumber := "?"
	if ho.PacketNumber != nil {
		version = fmt.Sprintf("%#d", *ho.PacketNumber)
	}
	return fmt.Sprintf("%#v,%#v", packetNumber, version)
}
// INITIAL(25,0xff00001d)[ACK,CRYPTO]
// Is represented as:
// PacketType: Initial
// HeaderOptions: HeaderOptions{ PacketNumber: 25, QUICVersion: 0xff00001d }
// frames: [ qt.AckFrame, qt.CryptoFrame ]
type AbstractSymbol struct {
	PacketType    qt.PacketType
	HeaderOptions HeaderOptions
	FrameTypes    mapset.Set // type: qt.FrameType
}

func (as *AbstractSymbol) String() string {
	packetType := packetTypeToString[as.PacketType]
	headerOptions := as.HeaderOptions.String()
	frameStrings := []string{}
	for _, frameType := range as.FrameTypes.ToSlice() {
		frameStrings = append(frameStrings, frameType.(qt.FrameType).String())
	}
	sort.Strings(frameStrings)
	frameTypes := strings.Join(frameStrings, ",")
	return fmt.Sprintf("%v(%v)[%v]", packetType, headerOptions, frameTypes)
}

func NewAbstractSymbol(packetType qt.PacketType, headerOptions HeaderOptions, frameTypes mapset.Set) AbstractSymbol {
	return AbstractSymbol{
		PacketType:    packetType,
		HeaderOptions: headerOptions,
		FrameTypes:    frameTypes,
	}
}

func NewAbstractSymbolFromString(message string) AbstractSymbol {
	messageStringRegex := regexp.MustCompile(`^([A-Z]+)(\(([0-9A-Za-zx,?]+)\))?\[([A-Z,]+)\]$`)
	subgroups := messageStringRegex.FindStringSubmatch(message)
	// The GetPacketType is the second group, we can get the type with a map.
	packetType := stringToPacketType[subgroups[1]]

	// GetHeader options contain options that might be optional, SHORT packets for example don't have QUICVersion.
	headerOptions := HeaderOptions{}
	// The fourth group has the content of header options.
	if subgroups[3] != "" {
		// We anticipate there might be more, so we split the string.
		headerOptionSlice := strings.Split(subgroups[3], ",")
		// The first option is the Packet Number.
		if headerOptionSlice[0] != "?" {
			parsedPacketNumber, err := strconv.ParseUint(headerOptionSlice[0], 10, 64)
			if err == nil {
				packetNumber := qt.PacketNumber(parsedPacketNumber)
				headerOptions.PacketNumber = &packetNumber
			}
		}

		// The second option is the QUIC Version.
		if headerOptionSlice[1] != "?" {
			parsedVersion, err := strconv.ParseUint(headerOptionSlice[1][2:], 16, 32)
			if err == nil {
				version32 := uint32(parsedVersion)
				headerOptions.QUICVersion = &version32
			}
		}
	}

	// The fifth group will be a CSV of frame types.
	frameTypes := mapset.NewSet()
	frameSplice := strings.Split(subgroups[4], ",")
	for _, frameString := range frameSplice {
		frameTypes.Add(qt.FrameTypeFromString(frameString))
	}

	return NewAbstractSymbol(packetType, headerOptions, frameTypes)
}

type AbstractSet struct {
	SymbolSet mapset.Set // type: AbstractSymbol
}

func NewAbstractSet() *AbstractSet {
	as := AbstractSet{SymbolSet: mapset.NewSet()}
	return &as
}

func (as *AbstractSet) Add(abstractSymbol AbstractSymbol) {
	as.SymbolSet.Add(abstractSymbol)
}

func (as *AbstractSet) Clear() {
	as.SymbolSet.Clear()
}

func (as *AbstractSet) String() string {
	if as.SymbolSet.Cardinality() == 0 {
		return "{}"
	}

	setSlice := as.SymbolSet.ToSlice()
	stringSlice := []string{}
	for _, setElement := range setSlice {
		symbol := setElement.(AbstractSymbol)
		stringSlice = append(stringSlice, (&symbol).String())
	}
	sort.Strings(stringSlice)

	return fmt.Sprintf("{%v}", strings.Join(stringSlice, ","))
}

type AbstractOrderedPair struct {
	AbstractInputs  []AbstractSymbol
	AbstractOutputs []AbstractSet
}

func (ct *AbstractOrderedPair) Input() *[]AbstractSymbol {
	return &ct.AbstractInputs
}

func (ct *AbstractOrderedPair) Output() *[]AbstractSet {
	return &ct.AbstractOutputs
}

func (ct *AbstractOrderedPair) SetInput(abstractSymbols []AbstractSymbol) {
	(*ct).AbstractInputs = abstractSymbols
}

func (ct *AbstractOrderedPair) SetOutput(abstractSets []AbstractSet) {
	(*ct).AbstractOutputs = abstractSets
}

func (ct *AbstractOrderedPair) String() string {
	aiStringSlice := []string{}
	for _, value := range ct.AbstractInputs {
		aiStringSlice = append(aiStringSlice, value.String())
	}
	aiString := fmt.Sprintf("[%v]", strings.Join(aiStringSlice, ","))

	aoStringSlice := []string{}
	for _, value := range ct.AbstractOutputs {
		aoStringSlice = append(aoStringSlice, value.String())
	}
	aoString := fmt.Sprintf("[%v]", strings.Join(aoStringSlice, ","))
	return fmt.Sprintf("(%v,%v)", aiString, aoString)
}

