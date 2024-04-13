package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() { // берем адрес сервера из аргументов командной строки
	conn, err := net.Dial("tcp", "localhost:8080") // открываем TCP-соединение к серверу
	if err != nil {
		log.Fatal(err)
	}

	go copyTo(os.Stdout, conn) // читаем из сокета в stdout
	copyTo(conn, os.Stdin)     // пишем в сокет из stdin
}

func copyTo(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
