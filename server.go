package main

import (
	"fmt"
	"github.com/colemickens/gobble"
	common "github.com/colemickens/goxpn/xpncommon"
	"log"
	"net"
	"net/http"
)

var lastUserId int = 0
var users [int]*user

func server_init() {
	users = make(map[int]*user)
}

func server() {
	server_init()

	addr, err := net.ResolveTCPAddr("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			lastUserId++

			log.Println("accept...")
			conn, err := conn.Accept()

			if err != nil {
				log.Println(err)
			}

			u := &user{
				id:          lastUserId,
				conn:        conn,
				transmitter: gobble.NewTransmitter(conn),
				receiver:    gobble.NewReceiver(conn),
			}

			users = append(users, u)
			go func() {
				for {
					msg, err := u.receiver.Receive()

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
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./ui/")))
	log.Fatal(http.ListenAndServe(":80", nil))
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

func (u *user) String() string {
	i := u.id
	a := u.udpAddr.String()
	g := "[none]"
	if u.curGroup != nil {
		g = u.curGroup.name
	}
	return fmt.Sprintf("id:%d, udpAddr:%s, curGroup:%s", i, a, g)
}
