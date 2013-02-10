package main

import (
	"code.google.com/p/nat"
	"encoding/gob"
	log "fmt"
	"net"
	"time"
)

var peerConnections map[int]*PeerConn

func client_init() {
	peerConnections = make(map[int]*PeerConn)
}

func run_client(host string) error {
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

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for {
		var env Envelope
		err := decoder.Decode(&env)
		if err != nil {
			log.Println("lost conn to server")
			// TODO: Return, let main try to reconnect to server, still
			return nil
		}

		if env.PcSignal != nil {
			HandlePcSignal(encoder, *env.PcSignal)
		}

		if env.UserList != nil {
			// This is the list of existing users.
			// Let's try to establish a PeerConn to each
			for _, u := range env.UserList {
				InitPeerConn(encoder, u.Id)
			}
		}
	}

	return nil
}

type ShimConn struct {
	encoder  *gob.Encoder
	to       int
	readChan chan []byte
}

func newShimConn(encoder *gob.Encoder, to int) *ShimConn {
	return &ShimConn{encoder, to, make(chan []byte)}
}

func (sc *ShimConn) Write(bytes []byte) (int, error) {
	env := &Envelope{
		PcSignal: &PcSignal{
			To:      sc.to,
			Payload: bytes,
		},
	}
	err := sc.encoder.Encode(env)
	if err != nil {
		// TODO:  handle this error
	}
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

func MakePeerConn(encoder *gob.Encoder, peerId int, initiator bool) *PeerConn {
	pc := &PeerConn{
		sideband:   newShimConn(encoder, peerId),
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

func InitPeerConn(encoder *gob.Encoder, peerId int) {
	MakePeerConn(encoder, peerId, true)
}

func HandlePcSignal(encoder *gob.Encoder, signal PcSignal) {
	pc, ok := peerConnections[signal.From]
	if !ok {
		pc = MakePeerConn(encoder, signal.From, false)
	}
	pc.sideband.readChan <- signal.Payload
}

func handleRemoteUdp(pc *PeerConn) {
	data := make([]byte, 65535)
	for {
		n, err := pc.udpConn.Read(data)
		log.Println("read from remote udp")
		_ = n
		if err != nil {
			// blek
			log.Println("ERR: ", err)
		} else if !pc.ignorePkts {
			log.Println("udp packet", data[:n])
			log.Println("echoing")
			time.Sleep(3 * time.Second)
			pc.udpConn.Write([]byte{0x00, 0x02, 0x04, 0x06, 0x08})
		}
	}
}
