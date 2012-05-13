package main

import (
	"fmt"
	"github.com/colemickens/gobble"
	common "github.com/colemickens/goxpn/xpncommon"
	"log"
	"net"
	"net/http"
)

type PcSignal struct {
	From    int
	To      int
	Payload []byte
}

func main() {
	mode := os.Args[1]

	host := "goxpn.us.to:9000"

	if mode == "client" {
		err := client(host)
		if err != nil {
			panic(err)
		}
	} else if mode == "server" {
		err := server(host)
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("unknown action")
	}
}
