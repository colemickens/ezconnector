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
			// TODO: Return, let main try to reconnect to server, still
			return nil
		}

		switch msg.(type) {
		case PcSignal:
			signal := msg.(PcSignal)
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
	signal := &PcSignal{
		To:      sc.to,
		Payload: bytes,
	}
	transmitter.Transmit(signal)
	return len(bytes), nil
}

func (sc *ShimConn) Read(bytes []byte) (int, error) {
	tmp := <-sc.readChan
	n := copy(bytes, tmp)
	return n, nil
}

func (sc *ShimConn) LocalAddr() net.Addr                { return nil }
func (sc *ShimConn) Close() error                       { return nil }
func (sc *ShimConn) RemoteAddr() net.Addr               { return nil }
func (sc *ShimConn) SetDeadline(t time.Time) error      { return nil }
func (sc *ShimConn) SetReadDeadline(t time.Time) error  { return nil }
func (sc *ShimConn) SetWriteDeadline(t time.Time) error { return nil }

type PeerConn struct {
	sideband   *ShimConn
	initiator  bool
	udpConn    net.Conn
	ignorePkts bool
}

func MakePeerConn(peerId int, initiator bool) *PeerConn {
	pc := &PeerConn{
		sideband:   newShimConn(peerId),
		initiator:  initiator,
		udpConn:    nil,
		ignorePkts: true,
	}
	go func() {
		var err error
		pc.udpConn, err = nat.Connect(pc.sideband, pc.initiator)
		if err != nil {
			log.Println("err doing nat conn", err)
			// TODO REMOVE FROM MAP
		} else {
			log.Println("nat busted!")
			go func() {
				time.Sleep(2 * time.Second)
				pc.ignorePkts = false
				pc.udpConn.Write([]byte{0x00, 0x01, 0x02, 0x03})
			}()
			handleRemoteUdp(pc)
		}
	}()
	peerConnections[peerId] = pc
	return pc
}

func InitPeerConn(peerId int) {
	MakePeerConn(peerId, true)
}

func HandlePcSignal(signal PcSignal) {
	pc, ok := peerConnections[signal.From]
	if !ok {
		pc = MakePeerConn(signal.From, false)
	}
	pc.sideband.readChan <- signal.Payload
}

func handleRemoteUdp(pc *PeerConn) {
	//data := make([]byte, 65535)
	var data []byte
	for {
		n, err := pc.udpConn.Read(data)
		log.Println("read from remote udp")
		_ = n
		if err != nil {
			// blek
		} else if !pc.ignorePkts {
			log.Println("udp packet", data)
			log.Println("echoing")
			time.Sleep(3 * time.Second)
			pc.udpConn.Write([]byte{0x00, 0x02, 0x04, 0x06, 0x08})
		}
	}
}
