package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering     = make(chan client)
	leaving      = make(chan client)
	messages     = make(chan string)
	lst_messages = make(chan string)
	pvt_messages = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados
	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	apelido := conn.RemoteAddr().String()
	ch <- "vc é " + apelido
	messages <- apelido + " chegou!"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		args := strings.Split(input.Text(), " ")
		cmd := args[0]

		switch cmd {
		case "/nick":
			fmt.Println(args)
			apelido = args[1]

		case "/quit":
			messages <- apelido + " se foi "
			leaving <- ch
			conn.Close()
		default:
			messages <- apelido + ":" + input.Text()
		}
	}

	leaving <- ch
	messages <- apelido + " se foi "
	conn.Close()
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
