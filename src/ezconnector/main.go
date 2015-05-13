package main

import (
	"encoding/gob"
	"flag"
	"net"
)

// An envelope may contain either
// There's nothing keeping it from being both
// But we'll never be sending a peer list except for
// on initial connect, and after that it will only be
// PcSignals
type Envelope struct {
	PcSignal *PcSignal
	UserList []User
}

type PcSignal struct {
	From    int
	To      int
	Payload []byte
}

type User struct {
	Id int

	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

var (
	server_flag = flag.String("server", "", "Port/interface to listen on")
	client_flag = flag.String("client", "", "Host/port to connect to")
)

func init() {
	flag.Parse()
}

func main() {
	if *server_flag != "" && *client_flag != "" {
		panic("Can't be both the server and a client")
	}

	if *server_flag != "" {
		err := run_server(*server_flag)
		if err != nil {
			panic(err)
		}
	} else if *client_flag != "" {
		err := run_client(*client_flag)
		if err != nil {
			panic(err)
		}
	} else {
		panic("Must either be client or server")
	}
}
