package adapter

import (
	"encoding/json"
	"fmt"
	mapset "github.com/tiferrei/golang-set"
	qt "github.com/tiferrei/quic-tracker"
	"github.com/tiferrei/quic-tracker/agents"
	tcp "github.com/tiferrei/tcp_server"
	"log"
	"os"
    "os/exec"
    "strings"
	"time"
)

type Adapter struct {
	connection             *qt.Connection
	trace                  *qt.Trace
	pcap                   *exec.Cmd
	http3                  bool
	httpPath               string
	waitTime               time.Duration
	agents                 *agents.ConnectionAgents
	server                 *tcp.Server
	stop                   chan bool
	Logger                 *log.Logger

	incomingLearnerSymbols qt.Broadcaster // Type: AbstractSymbol
	incomingSulPackets     chan interface{}
	outgoingSulPackets     chan interface{}
	outgoingPacket         *ConcreteSymbol
	incomingPacketSet      ConcreteSet
	incomingRequest        AbstractSymbol
	outgoingResponse       AbstractSet
	oracleTable            AbstractConcreteMap
}

func NewAdapter(adapterAddress string, sulAddress string, sulName string, http3 bool, httpPath string, tracing bool, waitTime time.Duration) (*Adapter, error) {
	adapter := new(Adapter)

	adapter.Logger = log.New(os.Stderr, "[ADAPTER] ", log.Lshortfile)
	adapter.Logger.Printf("Adapter Address: %v", adapterAddress)
	adapter.Logger.Printf("SUL Address: %v", sulAddress)
	adapter.Logger.Printf("SUL Name: %v", sulName)
	adapter.Logger.Printf("HTTP3: %v", http3)
	adapter.Logger.Printf("HTTP Path: %v", httpPath)
	adapter.Logger.Printf("TRACING: %v", tracing)
	adapter.Logger.Printf("Wait Time: %v", waitTime)

	adapter.incomingLearnerSymbols = qt.NewBroadcaster(1000)
	adapter.httpPath = httpPath
	adapter.http3 = http3
	adapter.waitTime = waitTime
	adapter.stop = make(chan bool, 1)
	adapter.server = tcp.New(adapterAddress)

	adapter.connection, _ = qt.NewDefaultConnection(sulAddress, sulName, nil, false, "hq", adapter.http3)
	if tracing {
	    var err error
        adapter.pcap, err = qt.StartPcapCapture(adapter.connection, "")
        if err != nil {
            panic(err)
        }

		adapter.trace = qt.NewTrace("Adapter", 1, sulAddress)
		adapter.trace.AttachTo(adapter.connection)

		adapter.trace.StartedAt = time.Now().Unix()
		ip := strings.Replace(adapter.connection.ConnectedIp().String(), "[", "", -1)
		adapter.trace.Ip = ip[:strings.LastIndex(ip, ":")]
	}

	adapter.incomingSulPackets = adapter.connection.IncomingPackets.RegisterNewChan(1000)
	adapter.outgoingSulPackets = adapter.connection.OutgoingPackets.RegisterNewChan(1000)

	adapter.outgoingPacket = nil
	adapter.incomingPacketSet = *NewConcreteSet()
	adapter.outgoingResponse = *NewAbstractSet()
	adapter.oracleTable = *NewAbstractConcreteMap()

	adapter.connection.TLSTPHandler.MaxStreamDataBidiLocal = 80

	adapter.agents = agents.AttachAgentsToConnection(adapter.connection, agents.GetBasicAgents()...)
	adapter.agents.Get("ClosingAgent").(*agents.ClosingAgent).WaitForFirstPacket = true
	adapter.agents.Add(&agents.HandshakeAgent{
		TLSAgent: adapter.agents.Get("TLSAgent").(*agents.TLSAgent),
		SocketAgent: adapter.agents.Get("SocketAgent").(*agents.SocketAgent),
		DisableFrameSending: true,
	})
	adapter.agents.Add(&agents.SendingAgent{
		MTU: 1200,
		FrameProducer: adapter.agents.GetFrameProducingAgents(),
	})
	adapter.agents.Get("StreamAgent").(*agents.StreamAgent).DisableFrameSending = true
	if adapter.http3 {
		adapter.agents.Add(&agents.HTTP3Agent{})
	} else {
		adapter.agents.Add(&agents.HTTP09Agent{})
	}
	adapter.agents.Get("SendingAgent").(*agents.SendingAgent).KeepDroppedEncryptionLevels = true
	adapter.agents.Get("FlowControlAgent").(*agents.FlowControlAgent).DisableFrameSending = true
	adapter.agents.Get("FlowControlAgent").(*agents.FlowControlAgent).DontSlideCreditWindow = true
	adapter.agents.Get("TLSAgent").(*agents.TLSAgent).DisableFrameSending = true
	adapter.agents.Get("AckAgent").(*agents.AckAgent).DisableAcks = map[qt.PNSpace]bool {
		qt.PNSpaceNoSpace: true,
		qt.PNSpaceInitial: true,
		qt.PNSpaceHandshake: true,
		qt.PNSpaceAppData: true,
	}

	adapter.server.OnNewMessage(adapter.handleNewServerInput)

	return adapter, nil
}

func (a *Adapter) Run() {
	go a.server.Listen()
	a.Logger.Printf("Server now listening.")
	incomingSymbolChannel := a.incomingLearnerSymbols.RegisterNewChan(1000)

	for {
		select {
		case i := <-incomingSymbolChannel:
			as := i.(AbstractSymbol)
			pnSpace := qt.PacketTypeToPNSpace[as.PacketType]
			encLevel := qt.PacketTypeToEncryptionLevel[as.PacketType]

			if as.HeaderOptions.QUICVersion != nil {
				a.connection.Version = *as.HeaderOptions.QUICVersion
			}

			if as.HeaderOptions.PacketNumber != nil {
				a.connection.PacketNumberLock.Lock()
				a.connection.PacketNumber[pnSpace] = *as.HeaderOptions.PacketNumber
				a.connection.PacketNumberLock.Unlock()
			}

			frameTypesSlice := []qt.FrameType{}
			for _, frameType := range as.FrameTypes.ToSlice() {
				frameTypesSlice = append(frameTypesSlice, frameType.(qt.FrameType))
			}
			for _, frameType := range frameTypesSlice {
				switch frameType {
				case qt.AckType:
					a.agents.Get("AckAgent").(*agents.AckAgent).SendFromQueue <- pnSpace
				case qt.PingType:
					a.connection.FrameQueue.Submit(qt.QueuedFrame{Frame: new(qt.PingFrame), EncryptionLevel: encLevel})
				case qt.CryptoType:
					a.agents.Get("TLSAgent").(*agents.TLSAgent).SendFromQueue <- encLevel
				case qt.PaddingFrameType:
					a.connection.FrameQueue.Submit(qt.QueuedFrame{Frame: new(qt.PaddingFrame), EncryptionLevel: encLevel})
				case qt.StreamType:
					if len(a.connection.StreamQueue[qt.FrameRequest{FrameType: qt.StreamType, EncryptionLevel: qt.EncryptionLevel1RTT}]) == 0 {
						if a.http3 {
							a.agents.Get("HTTP3Agent").(*agents.HTTP3Agent).SendRequest(a.httpPath, "GET", "quic.tiferrei.com", nil)
						} else {
							a.agents.Get("HTTP09Agent").(*agents.HTTP09Agent).SendRequest(a.httpPath, "GET", "quic.tiferrei.com", nil)
						}
					}
					time.Sleep(1 * time.Millisecond)
					a.agents.Get("StreamAgent").(*agents.StreamAgent).SendFromQueue <- qt.FrameRequest{qt.StreamType, encLevel}
				case qt.MaxDataType:
				case qt.MaxStreamDataType:
					a.agents.Get("FlowControlAgent").(*agents.FlowControlAgent).SendFromQueue <- qt.FrameRequest{frameType, encLevel}
				case qt.HandshakeDoneType:
					a.connection.FrameQueue.Submit(qt.QueuedFrame{Frame: new(qt.HandshakeDoneFrame), EncryptionLevel: encLevel})
				default:
					panic(fmt.Sprintf("Error: Frame Type '%v' not implemented!", frameType))
				}
			}
			// FIXME: This ensures the request gets queued before packets are sent. I'm not proud of it but it works.
			time.Sleep(3 * time.Millisecond)
			a.Logger.Printf("Submitting request: %v", as.String())
			a.connection.PreparePacket.Submit(encLevel)
		case o := <-a.incomingSulPackets:
			var packetType qt.PacketType
			version := &a.connection.Version
			frameTypes := mapset.NewSet()

			switch packet := o.(type) {
			case *qt.VersionNegotiationPacket:
				packetType = qt.VersionNegotiation
				version = &packet.Version
			case *qt.RetryPacket:
				packetType = qt.Retry
				version = nil
			case *qt.StatelessResetPacket:
				packetType = qt.StatelessReset
				version = nil
			case qt.Framer:
				packetType = packet.GetHeader().GetPacketType()
				// TODO: GetFrames() might not return a deterministic order. Idk yet.
				for _, frame := range packet.GetFrames() {
					if frame.FrameType() != qt.PaddingFrameType {
						// We don't want to pass PADDINGs to the learner.
						frameTypes.Add(frame.FrameType())
					}
				}
				// A framer with no frames is a result of removing retransmitted ones.
				// FIXME: This could be more elegant.
				if frameTypes.Cardinality() == 0 {
					continue
				}
			default:
				panic(fmt.Sprintf("Error: Packet '%T' not implemented!", packet))
			}

			concreteSymbol := NewConcreteSymbol(o.(qt.Packet))
			a.incomingPacketSet.Add(concreteSymbol)
			var packetNumber *qt.PacketNumber = nil
			if concreteSymbol.Packet.GetHeader() != nil {
			    pn := concreteSymbol.Packet.GetHeader().GetPacketNumber()
                packetNumber = &pn
            }

			if a.incomingRequest.HeaderOptions.PacketNumber == nil {
				packetNumber = nil
			}

			if a.incomingRequest.HeaderOptions.QUICVersion == nil {
				version = nil
			}

			abstractSymbol := NewAbstractSymbol(
				packetType,
				HeaderOptions{QUICVersion: version, PacketNumber: packetNumber},
				frameTypes)
			a.Logger.Printf("Got response: %v", abstractSymbol.String())
			a.outgoingResponse.Add(abstractSymbol)
		case o := <- a.outgoingSulPackets:
			cs := NewConcreteSymbol(o.(qt.Packet))
			a.outgoingPacket = &cs
		case <-a.stop:
			return
		default:
			// Got nothing this time...
		}
	}
}

func (a *Adapter) Stop() {
	a.SaveOracleTable(fmt.Sprintf("oracleTable-%d.json", time.Now().Unix()))
	a.SaveTrace(fmt.Sprintf("trace-%d.json", time.Now().Unix()))
	a.agents.Stop("SendingAgent")
	a.agents.StopAll()
	a.stop <- true
}

func (a *Adapter) Reset(client *tcp.Client) {
	a.Logger.Print("Received RESET command")
	a.agents.Stop("SendingAgent")
	a.agents.StopAll()
	a.connection.Close()
	a.connection, _ = qt.NewDefaultConnection(a.connection.ConnectedIp().String(), a.connection.ServerName, nil, false, "hq", a.http3)
	if a.trace != nil {
		a.trace.AttachTo(a.connection)
	}
	a.incomingSulPackets = a.connection.IncomingPackets.RegisterNewChan(1000)
	a.outgoingSulPackets = a.connection.OutgoingPackets.RegisterNewChan(1000)
	a.outgoingPacket = nil
	a.incomingPacketSet = *NewConcreteSet()
	a.outgoingResponse = *NewAbstractSet()

	a.connection.TLSTPHandler.MaxStreamDataBidiLocal = 80

	a.agents = agents.AttachAgentsToConnection(a.connection, agents.GetBasicAgents()...)
	a.agents.Get("ClosingAgent").(*agents.ClosingAgent).WaitForFirstPacket = true
	a.agents.Add(&agents.HandshakeAgent{
		TLSAgent: a.agents.Get("TLSAgent").(*agents.TLSAgent),
		SocketAgent: a.agents.Get("SocketAgent").(*agents.SocketAgent),
		DisableFrameSending: true,
	})
	a.agents.Add(&agents.SendingAgent{
		MTU: 1200,
		FrameProducer: a.agents.GetFrameProducingAgents(),
	})
	a.agents.Get("StreamAgent").(*agents.StreamAgent).DisableFrameSending = true
	if a.http3 {
		a.agents.Add(&agents.HTTP3Agent{})
	} else {
		a.agents.Add(&agents.HTTP09Agent{})
	}
	a.agents.Get("SendingAgent").(*agents.SendingAgent).KeepDroppedEncryptionLevels = true
	a.agents.Get("FlowControlAgent").(*agents.FlowControlAgent).DontSlideCreditWindow = true
	a.agents.Get("FlowControlAgent").(*agents.FlowControlAgent).DisableFrameSending = true
	a.agents.Get("TLSAgent").(*agents.TLSAgent).DisableFrameSending = true
	a.agents.Get("AckAgent").(*agents.AckAgent).DisableAcks = map[qt.PNSpace]bool {
		qt.PNSpaceNoSpace: true,
		qt.PNSpaceInitial: true,
		qt.PNSpaceHandshake: true,
		qt.PNSpaceAppData: true,
	}

	a.Logger.Print("Finished RESET mechanism")
	err := client.Send("DONE\n")
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (a *Adapter) handleNewServerInput(client *tcp.Client, message string) {
	message = strings.TrimSuffix(message, "\n")
	message = strings.TrimSuffix(message, "\r")
	query := strings.Split(message, " ")
	a.Logger.Printf("Server input: %v", query)
	if len(query) == 1 {
		switch query[0] {
		case "START":
			go a.Run()
		case "RESET":
			a.Reset(client)
		case "STOP":
			a.Stop()
			_ = client.Close()
			os.Exit(0)
		default:
			a.handleNewAbstractQuery(client, query)
		}
	} else {
		a.handleNewAbstractQuery(client, query)
	}
}

func (a *Adapter) handleNewAbstractQuery(client *tcp.Client, query []string) {
	abstractInputs := []AbstractSymbol{}
	abstractOutputs := []AbstractSet{}
	concreteInputs := []*ConcreteSymbol{}
	concreteOutputs := []ConcreteSet{}
	for _, message := range query {
		a.outgoingResponse = *NewAbstractSet()
		a.incomingPacketSet = *NewConcreteSet()
		a.outgoingPacket = nil
		a.incomingRequest = NewAbstractSymbolFromString(message)
		abstractInputs = append(abstractInputs, a.incomingRequest)

		// If we don't have the requested encryption level, skip and return EMPTY.
		if a.connection.CryptoState(qt.PacketTypeToEncryptionLevel[a.incomingRequest.PacketType]) != nil {
			a.incomingLearnerSymbols.Submit(a.incomingRequest)
			time.Sleep(a.waitTime)
		} else {
			a.Logger.Printf("Unable to send packet at " + qt.PacketTypeToEncryptionLevel[a.incomingRequest.PacketType].String() + " EL.")
		}

		abstractOutputs = append(abstractOutputs, a.outgoingResponse)
		concreteInputs = append(concreteInputs, a.outgoingPacket)
		concreteOutputs = append(concreteOutputs, a.incomingPacketSet)

		// If we received a Retry, give the connection time to restart.
		if strings.Contains(a.outgoingResponse.String(), "RETRY") {
			time.Sleep(400 * time.Millisecond)
		}
	}

	a.oracleTable.AddIOs(abstractInputs, abstractOutputs, concreteInputs, concreteOutputs)

	aoStringSlice := []string{}
	for _, value := range abstractOutputs {
		aoStringSlice = append(aoStringSlice, value.String())
	}

	err := client.Send(strings.Join(aoStringSlice, " ") + "\n")
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func writeJson(filename string, object interface{}) {
	outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err == nil {
		content, err := json.Marshal(object)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}

		outFile.Write(content)
		outFile.Close()
	}
}

func (a *Adapter) SaveTrace(filename string) {
	if a.trace != nil {
        err := a.trace.AddPcap(a.connection, a.pcap)
        if err != nil {
            a.trace.Results["pcap_error"] = err.Error()
        }
		a.connection.QLog.Title = "QUIC Adapter Trace"
		a.connection.QLogTrace.Sort()
		a.trace.QLog = a.connection.QLog
		a.trace.Complete(a.connection)
		outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err == nil {
			content, err := json.Marshal(a.trace)
			if err == nil {
				outFile.Write(content)
				outFile.Close()
			}
		}
	}
}

func (a *Adapter) SaveOracleTable(filename string) {
	writeJson(filename, a.oracleTable)
    fmt.Printf("Combining oracle tables with:.\n")
    fmt.Printf("    mv oracleTable.json oracleTable-$(date +%s).json || true && find . -name 'oracleTable*' -exec jq -s -c add {} > oracleTable.json +;\n")
    fmt.Printf("This operation can be very time consuming (15 min / GB to be combined).\n")
    fmt.Printf("Exiting the program now and running it natively on your adapter results may be faster.\n")

    cmd := "mv oracleTable.json oracleTable-$(date +%s).json || true && find . -name 'oracleTable*' -exec jq -s -c add {} > oracleTable.json +;"
    _, err := exec.Command("sh","-c",cmd).Output()
    if err != nil {
        fmt.Printf(fmt.Sprintf("Failed to combine oracle tables: %v\n", err))
    } else {
        fmt.Printf("Finished combining oracle tables.\n")
    }
}
