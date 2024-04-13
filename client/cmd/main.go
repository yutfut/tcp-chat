package main

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"io"
	"log"
	"net"
	"os"
	"tcp-chat/sign"
)

func main() { // берем адрес сервера из аргументов командной строки
	conn, err := net.Dial("tcp", "localhost:8080") // открываем TCP-соединение к серверу
	if err != nil {
		log.Fatal(err)
	}

	clientPrivateKey, clientPublicKey, err := sign.GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate key pair: %s", err)
	}

	clientPublicKeyBytes, err := sign.PublicKeyToBytes(clientPublicKey)
	if err != nil {
		log.Fatalf("failed to convert public key to bytes: %s", err)
	}

	if _, err = conn.Write(clientPublicKeyBytes); err != nil {
		log.Fatalf("failed to send client public key to server: %s", err)
	}

	serverPublicKeyBytes := make([]byte, 8192)
	serverPublicKeyLen, err := conn.Read(serverPublicKeyBytes)
	if err != nil {
		log.Fatalf("failed to read server public key: %s", err)
	}

	serverPublicKey, err := sign.BytesToPublicKey(serverPublicKeyBytes[:serverPublicKeyLen])
	if err != nil {
		log.Fatalf("failed to convert server public key to bytes: %s", err)
	}

	go ReadFromConn(os.Stdout, conn, clientPrivateKey)
	WriteToConn(conn, os.Stdin, serverPublicKey)
}

func ReadFromConn(out io.Writer, in net.Conn, clientPublicKey *rsa.PrivateKey) {
	message := make([]byte, 512)

	for {
		if _, err := in.Read(message); err != nil {
			log.Fatalf("failed to read from client: %s", err)
		}

		encodeMessage, err := sign.DecryptWithPrivateKey(message, clientPublicKey)
		if err != nil {
			log.Fatalf("failed to decrypt message: %s", err)
		}

		if _, err := io.Copy(out, bytes.NewReader(encodeMessage)); err != nil {
			log.Fatal(err)
		}
	}
}

func WriteToConn(out net.Conn, in io.Reader, serverPublicKey *rsa.PublicKey) {
	message := make([]byte, 512)

	for {
		messageLen, err := bufio.NewReader(in).Read(message)
		if err != nil {
			log.Fatalf("failed to read from client: %s", err)
		}

		decodeMessage, err := sign.EncryptWithPublicKey(message[:messageLen], serverPublicKey)
		if err != nil {
			log.Fatalf("failed to encrypt message: %s", err)
		}

		if _, err = out.Write(decodeMessage); err != nil {
			log.Fatalf("failed to write to client: %s", err)
		}
	}
}
