package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"proof-of-gork/pkg/pow"
	"proof-of-gork/pkg/quote"
	"strings"
)

// lets just hardcode for simplicity
const (
	host = "pow_server"
	port = "5555"
)

var powServer *pow.Server

func init() {
	powServer = pow.NewServer(&pow.Server{
		Difficulty:  3,
		NonceLength: 10,
		Secret:      []byte("qwerty123456"),
	})
}

// i also don't care about timeouts and context
// that's just for simplicity, i guess it's ok
func main() {
	l, e := net.Listen("tcp", host+":"+port)
	if e != nil {
		fmt.Println("Listening error:", e.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening TCP on " + host + ":" + port)

	for {
		conn, e := l.Accept()
		if e != nil {
			fmt.Println("Accept error:", e.Error())
			continue
		}
		// if server will be killed all running goroutines will be lost
		// so graceful shutdown should be implemented, with waiting until
		// goroutines are finished
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	message, _ := bufio.NewReader(conn).ReadString('\n')
	response := getResponse(strings.TrimSpace(message))
	_, e := conn.Write([]byte(response + "\n"))
	if e != nil {
		fmt.Println("Error while writing to connection:", e.Error())
		return
	}
}

func getResponse(message string) string {
	messageParts := strings.SplitN(message, " ", 2)

	// gimme
	if len(messageParts) == 1 && messageParts[0] == "gimme" {
		response, e := powServer.Generate()
		if e != nil {
			return "Error while generating pow challenge"
		}
		return response
	}

	// gimme <pow response>
	if len(messageParts) == 2 && messageParts[0] == "gimme" {
		ok, e := powServer.Validate(messageParts[1])
		if e != nil {
			return "Error while validating pow challenge"
		}
		if !ok {
			return "invalid pow solution"
		}
		q, e := quote.Random()
		if e != nil {
			return e.Error()
		}
		return fmt.Sprintf("%s (%s)", q.Content, q.Author)
	}

	return "invalid message format. expected: gimme OR gimme <pow response>"
}
