package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	ID      int
	Message []byte
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080") // открываем слушающий сокет
	if err != nil {
		log.Fatal(err)
	}

	message := make(chan Message)

	clients := make(map[int]net.Conn)

	id := 1

	go broadcast(clients, message)

	for {
		conn, err := listener.Accept() // принимаем TCP-соединение от клиента и создаем новый сокет
		if err != nil {
			continue
		}

		clients[id] = conn

		fmt.Println(clients)

		go handleClient(conn, message, id)

		id++
	}

	close(message)
}

func broadcast(clients map[int]net.Conn, message chan Message) {
	for {
		msg, ok := <-message
		if !ok {
			return
		}

		for i, client := range clients {
			if i == msg.ID {
				continue
			}
			_, err := client.Write(msg.Message)
			if err != nil {
				log.Fatal()
			}
		}
	}
}

func handleClient(conn net.Conn, message chan Message, id int) {
	defer conn.Close()

	buf1 := make([]byte, 128)
	_, err := conn.Write([]byte("Hello, what's your name?\n"))
	if err != nil {
		log.Fatal(err)
	}

	readLen, err := conn.Read(buf1)
	if err != nil {
		log.Println(err)
	}

	username := append(buf1[:readLen-1], []byte(": ")...)

	buf := make([]byte, 128)
	for {
		readLen, err = conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		message <- Message{
			ID:      id,
			Message: append(username, buf[:readLen]...),
		}
	}
}
