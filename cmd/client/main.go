package main

import (
	"bufio"
	"net"
	"proof-of-gork/pkg/pow"
	"strings"
	"time"
)
import "fmt"

// lets just hardcode for simplicity
const (
	host = "pow_server"
	port = "5555"
)

// i don't check server response format just for simplicity
// of course in real app this is must have
// also code is a little bit messy but i think it's of for prototype
func main() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for range t.C {
		// func here dut to defer usage
		func() {
			// first tcp connection
			conn1, e := net.Dial("tcp", host+":"+port)
			if e != nil {
				fmt.Printf("ERROR can't connect to server: %s \n", e.Error())
				return
			}
			defer conn1.Close()
			reader := bufio.NewReader(conn1)

			fmt.Println("sending gimme request...")
			c := []byte("gimme\n")
			if _, e := conn1.Write(c); e != nil {
				fmt.Printf("ERROR while sending message to server: %s \n", e.Error())
				return
			}
			chRaw, e := reader.ReadString('\n')
			if e != nil {
				fmt.Printf("ERROR while reading from server: %s \n", e.Error())
				return
			}
			chRaw = strings.TrimSpace(chRaw)
			fmt.Printf("pow challenge: %s\n", chRaw)
			solution, hash, duration, e := solveChallenge(chRaw)
			fmt.Printf("solution found: %s\n", solution)
			fmt.Printf("hash : %s\n", hash)
			fmt.Printf("time spent: %s\n", duration)

			// second tcp connection
			conn2, _ := net.Dial("tcp", host+":"+port)
			if e != nil {
				fmt.Printf("ERROR can't connect to server: %s \n", e.Error())
				return
			}
			defer conn2.Close()
			reader = bufio.NewReader(conn2)

			c = []byte(fmt.Sprintf("gimme %s\n", solution))
			if _, e := conn2.Write(c); e != nil {
				fmt.Printf("ERROR while sending message to server: %s \n", e.Error())
				return
			}

			quote, e := reader.ReadString('\n')
			if e != nil {
				fmt.Printf("ERROR while reading from server: %s \n", e.Error())
				return
			}
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println(strings.TrimSpace(quote))
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			fmt.Println()
		}()
	}
}

func solveChallenge(chRaw string) (string, []byte, time.Duration, error) {
	ch, e := pow.NewChallenge(chRaw)
	if e != nil {
		return "", nil, 0, fmt.Errorf("ERROR while parsing challenge: %s \n", e.Error())
	}

	start := time.Now()
	solution, hash, e := pow.Solve(ch, pow.DefaultHashStrategy)
	duration := time.Since(start)

	if e != nil {
		return "", nil, 0, fmt.Errorf("ERROR while solving challenge: %s \n", e.Error())
	}
	return solution, hash, duration, nil
}
