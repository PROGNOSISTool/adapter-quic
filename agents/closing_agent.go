package agents

import (
	"time"

	. "github.com/PROGNOSISTool/adapter-quic"
)

// The ClosingAgent is responsible for keeping track of events that can close the connection, such as the idle timeout.
// It can queue an (CONNECTION|APPLICATION)_CLOSE frame and wait for it to be sent out.
type ClosingAgent struct {
	BaseAgent
	closing             bool
	conn                *Connection
	IdleDuration        time.Duration
	IdleTimeout         *time.Timer
	WaitForFirstPacket  bool
	receivedFirstPacket bool
}

func (a *ClosingAgent) Run(conn *Connection) {  // TODO: Observe incoming CC and AC
	a.Init("ClosingAgent", conn.OriginalDestinationCID)
	a.conn = conn
	a.IdleDuration = time.Duration(a.conn.TLSTPHandler.IdleTimeout) * time.Millisecond
	a.IdleTimeout = time.NewTimer(a.IdleDuration)

	incomingPackets := conn.IncomingPackets.RegisterNewChan(1000)
	outgoingPackets := conn.OutgoingPackets.RegisterNewChan(1000)

	go func() {
		defer a.Logger.Println("Agent terminated")
		defer close(a.closed)

		for {
			select {
			case <-incomingPackets:
				a.receivedFirstPacket = true
				a.IdleTimeout.Reset(a.IdleDuration)
			case i := <-outgoingPackets:
				switch p := i.(type) {
				case Framer:
					if a.closing && (p.Contains(ConnectionCloseType) || p.Contains(ApplicationCloseType)) {
						close(a.conn.ConnectionClosed)
						return
					}
				}
				if p := i.(Packet); p.ShouldBeAcknowledged() {
					a.IdleTimeout.Reset(a.IdleDuration)
				}
			case <-a.IdleTimeout.C:
				if !a.WaitForFirstPacket || a.receivedFirstPacket {
					a.closing = true
					a.Logger.Printf("Idle timeout of %v reached, closing\n", a.IdleDuration.String())
					close(a.conn.ConnectionClosed)
					return
				}
			case shouldRestart := <-a.close:
				if !shouldRestart {
					close(a.conn.ConnectionClosed)
				}
				return
			}
		}
	}()
}

func (a *ClosingAgent) Close(quicLayer bool, errorCode uint64, reasonPhrase string) {
	if !a.closing {
		a.closing = true
		a.conn.CloseConnection(quicLayer, errorCode, reasonPhrase)
	}
}
