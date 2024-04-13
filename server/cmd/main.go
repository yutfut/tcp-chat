package main

import (
	"crypto/rsa"
	"log"
	"net"
	"tcp-chat/sign"
)

type Message struct {
	ID      int
	Message []byte
}

type Client struct {
	ClientConn      net.Conn
	ClientPublicKey *rsa.PublicKey
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverPrivateKey, serverPublicKey, err := sign.GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate key pair: %v", err)
	}

	serverPublicKeyByte, err := sign.PublicKeyToBytes(serverPublicKey)
	if err != nil {
		log.Fatalf("failed to convert public key to bytes: %v", err)
	}

	message := make(chan Message)
	clients := make(map[int]Client)
	id := 1

	go broadcast(clients, message)

	clientPublicKeyBytes := make([]byte, 2048)
	for {
		var conn net.Conn
		if conn, err = listener.Accept(); err != nil {
			continue
		}

		if _, err = conn.Read(clientPublicKeyBytes); err != nil {
			log.Printf("Error reading from client: %v", err)
		}

		if _, err = conn.Write(serverPublicKeyByte); err != nil {
			log.Printf("Error writing to client: %v", err)
		}

		clientPublicKey, err := sign.BytesToPublicKey(clientPublicKeyBytes)
		if err != nil {
			log.Printf("Error converting bytes to public key: %v", err)
		}

		clients[id] = Client{
			ClientConn:      conn,
			ClientPublicKey: clientPublicKey,
		}

		go handleClient(conn, message, id, clientPublicKey, serverPrivateKey)

		id++
	}

	close(message)
}

func broadcast(clients map[int]Client, message chan Message) {
	for {
		msg, ok := <-message
		if !ok {
			return
		}

		for i, client := range clients {
			if i == msg.ID {
				continue
			}

			decryptMessage, err := sign.EncryptWithPublicKey(msg.Message, client.ClientPublicKey)
			if err != nil {
				log.Fatalf("server encrypt error: %v", err)
			}

			if _, err = client.ClientConn.Write(decryptMessage); err != nil {
				log.Printf("server writing error: %v", err)
			}
		}
	}
}

func handleClient(conn net.Conn, message chan Message, id int, clientPrivateKey *rsa.PublicKey, serverPrivateKey *rsa.PrivateKey) {
	defer conn.Close()

	encryptMessage, err := sign.EncryptWithPublicKey([]byte("Hello, what's your name?\n"), clientPrivateKey)
	if err != nil {
		log.Printf("server encrypt error: %v", err)
		return
	}

	if _, err = conn.Write(encryptMessage); err != nil {
		log.Printf("server writing error: %v", err)
		return
	}

	buf := make([]byte, 512)
	if _, err = conn.Read(buf); err != nil {
		log.Printf("Error reading from server: %v", err)
		return
	}

	decryptUsername, err := sign.DecryptWithPrivateKey(buf, serverPrivateKey)
	if err != nil {
		log.Printf("server decrypt error: %v", err)
		return
	}
	username := append(decryptUsername[:len(decryptUsername)-1], []byte(": ")...)

	for {
		if _, err = conn.Read(buf); err != nil {
			log.Printf("server reading error: %v", err)
			return
		}

		encryptMessage, err = sign.DecryptWithPrivateKey(buf, serverPrivateKey)
		if err != nil {
			log.Printf("server decrypt error: %v", err)
			return
		}

		message <- Message{
			ID:      id,
			Message: append(username, encryptMessage...),
		}
	}
}
