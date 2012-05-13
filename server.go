package main

import (
	"github.com/colemickens/gobble"
	common "github.com/colemickens/goxpn/xpncommon"
	"log"
	"net"
	"net/http"
)

var lastUserId int = 0
var users map[int]*user

func server_init() {
	users = make(map[int]*user)
}

func server(host string) error {
	server_init()

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}

	conn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		for {
			lastUserId++

			log.Println("accept...")
			conn, err := conn.Accept()

			if err != nil {
				log.Println("err accepting", err)
			}

			u := &user{
				id:          lastUserId,
				conn:        conn,
				transmitter: gobble.NewTransmitter(conn),
				receiver:    gobble.NewReceiver(conn),
			}

			users[lastUserId] = u
			go func() {
				for {
					msg, _ := u.receiver.Receive()

					switch msg.(type) {

					case common.PcSignal:
						s := msg.(common.PcSignal)
						s.From = u.id

						log.Println("PcSignal from", s.From, "to", s.To, ":", s)

						toUser := userById(s.To)
						if toUser != nil {
							toUser.transmitter.Transmit(s)
						}
					}
				}
			}()

			// tell about other user(s)
			for id, user := range users {
				if id != u.id {
					u.transmitter.Transmit(id)
				}
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./ui/")))
	log.Fatal(http.ListenAndServe(":80", nil))

	return nil
}

type user struct {
	id int

	conn        net.Conn
	transmitter *gobble.Transmitter
	receiver    *gobble.Receiver
}

func userById(id int) *user {
	for _, u := range users {
		if u.id == id {
			return u
		}
	}
	return nil
}
