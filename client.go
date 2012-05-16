package main

import (
	"code.google.com/p/nat"
	log "fmt"
	"github.com/colemickens/gobble"
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
			log.Println("lost conn to server")
			return nil
		}

		switch msg.(type) {
		case PcSignal:
			signal := msg.(PcSignal)
			log.Println("client.receiver.Receiver", signal)
			HandlePcSignal(signal)
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
	log.Println("Writing via Write(", len(bytes), ")", bytes)
	signal := &PcSignal{
		To:      sc.to,
		Payload: bytes,
	}
	log.Println("client.transmitter.Transmit", signal)
	transmitter.Transmit(signal)
	return len(bytes), nil
}

func (sc *ShimConn) Read(bytes []byte) (int, error) {
	tmp := <-sc.readChan
	n := copy(bytes, tmp)
	log.Println("Reading via Read(", n, ")", bytes[:n])
	return n, nil
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
			log.Println("nat busted bitch")
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

func HandlePcSignal(signal PcSignal) {
	pc, ok := peerConnections[signal.From]
	if !ok {
		pc = MakePeerConn(signal.From, false)
	}
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
			log.Println(data)
			log.Println(string(data))
		}
	}
}
