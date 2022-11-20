package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string

var (
	entering     = make(chan client)
	leaving      = make(chan client)
	messages     = make(chan string)
	users        = make(map[string]client)
	pvt_messages = make(chan string)
	bot          = make(chan client)
	bot_msg      = make(chan string)
)

func Reverse_text(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

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

		case msg := <-pvt_messages:
			// broadcast de mensagens. Envio para todos
			arg := strings.Split(msg, " ")
			println("ENVIADO PARA", arg[3])
			cliente := users[arg[3]]
			cliente <- msg

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
	bot := make(chan string)
	go clientWriter(conn, ch)
	go clientWriter(conn, bot)
	users["bot"] = bot

	apelido := conn.RemoteAddr().String()
	ch <- "vc é " + apelido
	messages <- apelido + " chegou!"
	entering <- ch
	users[apelido] = ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		args := strings.Split(input.Text(), " ")
		cmd := args[0]

		switch cmd {
		case "/nick":
			if input.Text() == "/nick" {
				messages <- "NICK INVÁLIDO"
			} else {
				fmt.Println(args)
				delete(users, apelido)
				apelido = args[1]
				users[apelido] = ch
			}

		case "/quit":
			fmt.Println(args)
			leaving <- ch
			messages <- apelido + " se foi "
			delete(users, apelido)
			return

		case "/list":
			fmt.Println(args)
			for usuarios, nick := range users {
				fmt.Println("USUARIO ", usuarios, " nick ", nick)
				pvt_messages <- "Servidor " + " para " + apelido + " Usuários online: " + usuarios
			}

		case "/pvt":
			fmt.Println(args)
			mensagem := strings.SplitAfter(input.Text(), args[1])
			pvt_messages <- apelido + " enviando para " + args[1] + " " + mensagem[1]

		case "/bot":
			fmt.Println(args)
			mensagem := strings.SplitAfter(input.Text(), "/bot")
			pvt_messages <- "BOT" + " enviando para " + apelido + " : " + Reverse_text(mensagem[1])

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
