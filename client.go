package main

import (
	"code.google.com/p/nat"
	"github.com/colemickens/gobble"
	"log"
	"net"
	"time"
)

var peerConnections map[int]*PeerConn
var transmitter *gobble.Transmitter

func client_init() {
	peerConnections = make(map[int]*PeerConn)
}

func client(host string) error {
	client_init()

	var err error

	log.Println("trying to resolve server:", host)

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}

	log.Println("trying to connect to server:", addr)
	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		return err
	}

	transmitter = gobble.NewTransmitter(conn)
	receiver := gobble.NewReceiver(conn)

	for {
		msg, err := receiver.Receive()

		if err != nil {
			return err
		}

		switch msg.(type) {
		case PcSignal:
			signal := msg.(PcSignal)
			HandlePcSignal(&signal)
		case int:
			// this is a previously connected client that
			// the server is encouraging us to connect to
			peerId := msg.(int)
			InitPeerConn(peerId)
		}
	}

	return nil
}

type ShimConn struct {
	to       int
	readChan chan []byte
}

func newShimConn(to int) *ShimConn {
	return &ShimConn{to, make(chan []byte)}
}

func (sc *ShimConn) Write(bytes []byte) (n int, err error) {
	signal := &PcSignal{
		To:      sc.to,
		Payload: bytes,
	}
	log.Println("client.transmitter.Transmit To", signal.To)
	transmitter.Transmit(signal)
	return len(bytes), nil
}

func (sc *ShimConn) Read(bytes []byte) (n int, err error) {
	bytes = <-sc.readChan
	return len(bytes), nil
}

func (sc *ShimConn) LocalAddr() net.Addr                { return nil }
func (sc *ShimConn) Close() error                       { return nil }
func (sc *ShimConn) RemoteAddr() net.Addr               { return nil }
func (sc *ShimConn) SetDeadline(t time.Time) error      { return nil }
func (sc *ShimConn) SetReadDeadline(t time.Time) error  { return nil }
func (sc *ShimConn) SetWriteDeadline(t time.Time) error { return nil }

type PeerConn struct {
	sideband  *ShimConn
	initiator bool
	udpConn   net.Conn
}

func MakePeerConn(peerId int, initiator bool) *PeerConn {
	pc := &PeerConn{
		sideband:  newShimConn(peerId),
		initiator: initiator,
		udpConn:   nil,
	}
	go func() {
		var err error
		pc.udpConn, err = nat.Connect(pc.sideband, pc.initiator)
		if err != nil {
			log.Println("err doing nat conn", err)
			log.Println("(remove from map?)")
		} else {
			handleRemoteUdp(&pc.udpConn)
		}
	}()
	peerConnections[peerId] = pc
	return pc
}

func InitPeerConn(peerId int) {
	log.Println("InitPeerConn(", peerId, ")")
	MakePeerConn(peerId, true)
}

func HandlePcSignal(signal *PcSignal) {
	log.Println("HandlePcSignal(", signal, ")")
	pc, ok := peerConnections[signal.From]
	if !ok {
		log.Println("Created new peer conn")
		pc = MakePeerConn(signal.From, false)
	}
	log.Println("handing off payload")
	pc.sideband.readChan <- signal.Payload
}

func handleRemoteUdp(conn *net.Conn) {
	data := make([]byte, 65535)
	for {
		n, err := (*conn).Read(data)
		log.Println("read from remote udp")
		_ = n
		if err != nil {
			// blek
		} else {
			//packet := ParseRemoteUdpPacket(data)
			//_ = packet
			//handle.Inject(packet.AsTransmittablePcap())
			log.Println("recvd packet properly")
		}
	}
}
