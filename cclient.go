package main

import (
	. "./classes"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <port>\n", os.Args[0])
		os.Exit(-1)
	}

	conn, err := net.Dial("tcp", os.Args[1])

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	client := CreateClient(conn, nil)
	fmt.Printf("%v", conn.LocalAddr())
	go func() {
		for {
			out.WriteString(string(client.GetIncoming()) + "\n")
			out.Flush()
		}
	}()

	for {
		line, _, _ := in.ReadLine()
		client.PutOutgoing([]byte(string(line)))
	}

}
