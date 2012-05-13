package main

import (
	"log"
	"os"
)

type PcSignal struct {
	From    int
	To      int
	Payload []byte
}

func main() {
	mode := os.Args[1]

	shost := ":9000"
	chost := "goxpn.us.to" + shost

	if mode == "client" {
		err := client(chost)
		if err != nil {
			panic(err)
		}
	} else if mode == "server" {
		err := server(shost)
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("unknown action")
	}
}
