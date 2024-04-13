package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"tcp-chat/sign"
)

type Message struct {
	ID      int
	Message []byte
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080") // открываем слушающий сокет
	if err != nil {
		log.Fatal(1, err)
	}

	serverPrivateKey, serverPublicKey := sign.GenerateKeyPair()

	fmt.Println(serverPrivateKey, serverPublicKey)

	serverPublicKeyByte := sign.PublicKeyToBytes(serverPublicKey)

	message := make(chan Message)

	clients := make(map[int]net.Conn)

	id := 1

	go broadcast(clients, message, serverPrivateKey)

	clientPublicKeyBytes := make([]byte, 8192)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		read, err := conn.Read(clientPublicKeyBytes)
		if err != nil {
			log.Fatal(2, err)
		}
		fmt.Println(read)

		_, err = conn.Write(serverPublicKeyByte)
		if err != nil {
			log.Fatal(3, err)
		}

		clientPublicKey := sign.BytesToPublicKey(clientPublicKeyBytes[:read])

		clients[id] = conn

		go handleClient(conn, message, id, clientPublicKey)

		id++
	}

	close(message)
}

func broadcast(clients map[int]net.Conn, message chan Message, serverPrivateKey *rsa.PrivateKey) {
	for {
		msg, ok := <-message
		if !ok {
			return
		}

		decryptMessage := sign.DecryptWithPrivateKey(msg.Message, serverPrivateKey)

		for i, client := range clients {
			if i == msg.ID {
				continue
			}
			_, err := client.Write(decryptMessage)
			if err != nil {
				log.Fatal(4, err)
			}
		}
	}
}

func handleClient(conn net.Conn, message chan Message, id int, clientPublicKey *rsa.PublicKey) {
	defer conn.Close()

	buf1 := make([]byte, 8192)

	//encryptMessage := sign.EncryptWithPublicKey([]byte("Hello, what's your name?\n"), clientPublicKey)
	encryptMessage := sign.EncryptWithPublicKey([]byte("Hello\n"), clientPublicKey)
	_, err := conn.Write(encryptMessage)
	if err != nil {
		log.Fatal(5, err)
	}

	readLen, err := conn.Read(buf1)
	if err != nil {
		log.Println(6, err)
	}

	username := append(buf1[:readLen-1], []byte(": ")...)

	buf := make([]byte, 8192)
	for {
		readLen, err = conn.Read(buf)
		if err != nil {
			log.Fatal(7, err)
		}

		encryptMessage = sign.EncryptWithPublicKey(buf[:readLen], clientPublicKey)

		message <- Message{
			ID:      id,
			Message: append(username, encryptMessage...),
		}
	}
}
