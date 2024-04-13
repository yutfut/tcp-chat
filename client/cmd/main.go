package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"tcp-chat/sign"
)

func main() { // берем адрес сервера из аргументов командной строки
	conn, err := net.Dial("tcp", "localhost:8080") // открываем TCP-соединение к серверу
	if err != nil {
		fmt.Println(1)
		log.Fatal(err)
	}

	clientPrivateKey, clientPublicKey := sign.GenerateKeyPair()

	fmt.Println("client keys", clientPrivateKey, clientPublicKey)

	clientPublicKeyBytes := sign.PublicKeyToBytes(clientPublicKey)

	fmt.Println("key for server", clientPublicKeyBytes)

	write, err := conn.Write(clientPublicKeyBytes)
	if err != nil {
		fmt.Println(2)
		log.Fatal(err)
	}

	fmt.Println(write)

	serverPublicKeyBytes := make([]byte, 8192)
	serverPublicKeyLen, err := conn.Read(serverPublicKeyBytes)
	if err != nil {
		fmt.Println(3)
		log.Fatal(err)
	}

	fmt.Println("server key", serverPublicKeyLen, serverPublicKeyBytes)

	serverPublicKey := sign.BytesToPublicKey(serverPublicKeyBytes[:serverPublicKeyLen])

	fmt.Println("server private key", serverPublicKey)

	go ReadFromConn(os.Stdout, conn, clientPrivateKey)
	WriteToConn(conn, os.Stdin, serverPublicKey)
}

func ReadFromConn(out io.Writer, in net.Conn, clientPublicKey *rsa.PrivateKey) {
	message := make([]byte, 8192)

	for {
		read, err := in.Read(message)
		if err != nil {
			fmt.Println(4)
			log.Fatal(err)
		}

		fmt.Println("read message", message[:read])

		encodeMessage := sign.DecryptWithPrivateKey(message[:read], clientPublicKey)

		fmt.Println("encode message", encodeMessage)
		if _, err = io.Copy(out, bytes.NewReader(encodeMessage)); err != nil {
			fmt.Println(5)
			log.Fatal(err)
		}
	}
}

func WriteToConn(out net.Conn, in io.Writer, serverPublicKey *rsa.PublicKey) {
	message := make([]byte, 8192)

	for {
		write, err := in.Write(message)
		if err != nil {
			fmt.Println(6)
			log.Fatal(err)
		}
		if message[0] == byte(0) {
			continue
		}

		decodeMessage := sign.EncryptWithPublicKey(message[:write], serverPublicKey)

		if _, err = io.Copy(out, bytes.NewReader(decodeMessage)); err != nil {
			fmt.Println(7)
			log.Fatal(err)
		}
	}
}

func copyTo(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
