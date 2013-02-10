package main

import (
	"encoding/gob"
	"log"
	"net"
)

var lastUserId int = 0
var users map[int]*User

func server_init() {
	users = make(map[int]*User)
}

func run_server(host string) error {
	server_init()

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}

	conn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	for {
		lastUserId++

		log.Println("accepting new user...")
		conn, err := conn.Accept()
		log.Println("accepted user:", lastUserId)

		if err != nil {
			log.Println("err accepting", err)
			continue
		}

		u := &User{
			Id:      lastUserId,
			conn:    conn,
			encoder: gob.NewEncoder(conn),
			decoder: gob.NewDecoder(conn),
		}

		users[lastUserId] = u
		go func() {
			for {
				env := new(Envelope)
				err := u.decoder.Decode(&env)
				if err != nil {
					log.Println("forgetting user: ", u.Id)
					// remove this user/conn from the map?
					return
				}

				if env.PcSignal != nil {
					env.PcSignal.From = u.Id

					log.Println("pcsignal", env.PcSignal.From, "->", env.PcSignal.To)

					toUser := users[env.PcSignal.To]
					if toUser != nil {
						toUser.encoder.Encode(env)
					}
				}
			}
		}()

		// You just connected, let's tell you about the other users and you can connect to them
		var userList []User
		for _, iter_user := range users {
			if iter_user.Id != u.Id { // ignore ourself (the client doesn't know their own ID because this is a trivial example app)
				userList = append(userList, *iter_user)
			}
		}
		env := &Envelope{
			UserList: userList,
		}
		err = u.encoder.Encode(env)
		if err != nil {
			// TODO: what do we do here
		}
	}

	return nil
}

/*
func userById(id int) *User {
	for _, u := range users {
		if u.Id == id {
			return u
		}
	}
	return nil
}
*/
