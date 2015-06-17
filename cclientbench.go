package main

import (
	. "./classes"
	//"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type (
	Clients []*Client
)

func startClient() (client *Client) {
	conn, err := net.Dial("tcp", os.Args[1])

	if err != nil {
		log.Fatal(err)
	}

	client = CreateClient(conn, nil)
	go readclient(client)
	return
}

func readclient(client *Client) {
	for {
		data := client.GetIncoming()
		fmt.Println(client.GetConn().RemoteAddr().String(), ":", string(data))
	}
}

func startClients(N int) (clients Clients) {
	clients = make(Clients, N)
	for i := 0; i < N; i++ {
		clients[i] = startClient()
	}
	return
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <port>\n", os.Args[0])
		os.Exit(-1)
	}
	exit := make(chan bool)
	cc := startClients(40000)

	for x, ss := range cc {
		func(ii int, sc *Client) {
			if ii%100 == 0 {
				sc.PutOutgoing([]byte(fmt.Sprintf("hello, i am conncect :%d", ii)))
				time.Sleep(time.Second * 1)
			}
			if x == 21001 {
				exit <- true
			}
		}(x, ss)
	}
	/*
		mapc := make(map[int]*Client, 32000)
		for i := 0; i < 32000; i++ {
			go func(x int) {
				conn, err := net.Dial("tcp", os.Args[1])

				if err != nil {
					log.Fatal(err)
				}

				in := bufio.NewReader(os.Stdin)
				out := bufio.NewWriter(os.Stdout)

				client := CreateClient(conn, nil)
				client.PutOutgoing([]byte(fmt.Sprintf("hello, i am conncect :%d", x)))
				fmt.Printf("%v", conn.LocalAddr())
				mapc[x] = client

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
			}(i)
			/*
				for {
					line, _, _ := in.ReadLine()
					client.PutOutgoing([]byte(string(line)))
				}
	*/
	//}
	<-exit
	//time.Sleep(time.Hour * 1)

}
