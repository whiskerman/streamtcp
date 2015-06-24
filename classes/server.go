package streamtcp

import (
	"fmt"
	"log"
	"net"
	"sync"
	//"strings"
)

const (
	MAXCLIENTS = 1000000
)

type CallBackServer func(net.Conn, []byte)

type Message chan []byte
type Token chan int
type ClientTable map[net.Conn]*Session

type Server struct {
	locker     *sync.RWMutex
	listener   net.Listener
	sessions   ClientTable
	tokens     Token
	pending    chan net.Conn
	quiting    chan net.Conn
	incoming   Message
	outgoing   Message
	messageRec CallBackServer
}

func (self *Server) generateToken() {
	self.tokens <- 0
}

func (self *Server) takeToken() {
	<-self.tokens
}

func CreateServer(callbackServer CallBackServer) *Server {
	server := &Server{
		sessions:   make(ClientTable, MAXCLIENTS),
		tokens:     make(Token, MAXCLIENTS),
		pending:    make(chan net.Conn, 1024000),
		quiting:    make(chan net.Conn, 1024000),
		incoming:   make(Message, 1024000),
		outgoing:   make(Message, 1024000),
		locker:     new(sync.RWMutex),
		messageRec: callbackServer,
	}
	//server.listeners = make([]net.Listener, 100)
	server.listen()
	return server
}

func (self *Server) listen() {
	go func() {
		for {
			select {
			case message := <-self.incoming:
				//go func(msg []byte) {
				self.broadcast(message)
			//	}(message)
			case conn := <-self.pending:
				//go func(con net.Conn) {
				self.join(conn)
				//}(con)
			case conn := <-self.quiting:
				//go func(con net.Conn) {
				self.leave(conn)
				//}(conn)
			}
		}
	}()
}

func (self *Server) join(conn net.Conn) {
	session := CreateSession(conn, nil)
	name := getUniqName()
	session.SetName(name)
	self.sessions[conn] = session

	log.Printf("Auto assigned name for conn %p: %s\n", conn, name)
	go func() {
		for {
			msg := <-session.incoming
			/*
				log.Printf("Got message: %s from client %s\n", msg, session.GetName())

				if strings.HasPrefix(msg, ":") {
					if cmd, err := parseCommand(msg); err == nil {
						if err = self.executeCommand(session, cmd); err == nil {
							continue
						} else {
							log.Println(err.Error())
						}
					} else {
						log.Println(err.Error())
					}
				}
			*/
			//fmt.Println(session.conn, " incoming: ", msg)
			if self.messageRec != nil {
				self.messageRec(conn, msg)
			}
			// fallthrough to normal message if it is not parsable or executable
			self.incoming <- []byte(fmt.Sprintf("%s says: %s", session.GetName(), string(msg)))
		}
	}()

	go func() {
		for {
			conn := <-session.quiting
			log.Printf("Client %s is quiting\n", session.GetName())
			self.quiting <- conn
		}
	}()
}

func (self *Server) leave(conn net.Conn) {
	self.locker.Lock()
	defer self.locker.Unlock()
	if conn != nil {
		//close(self.sessions[conn].incoming)
		//close(self.sessions[conn].outgoing)
		conn.Close()

		delete(self.sessions, conn)

	}

	self.generateToken()
}

func (self *Server) broadcast(message []byte) {
	//log.Printf("Broadcasting message: %s\n", message)
	//go func() {
	for _, client := range self.sessions {
		if !client.closing {
			client.outgoing <- message
		}

		//fmt.Println(client.conn, ":", message)
	}
	//}()
}

func (self *Server) Start(connString string) {

	var err error
	self.listener, err = net.Listen("tcp", connString)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Server %p starts%v\n  x is %d", self, self, 1)
	//exit := make(chan bool)
	// filling the tokens
	for i := 0; i < MAXCLIENTS; i++ {
		self.generateToken()
	}
	//for j := 0; j < 10; j++ {

	func(m int) {
		for {
			conn, err := self.listener.Accept()

			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("A new connection %v kicks\n", conn)

			self.takeToken()
			self.pending <- conn
		}
	}(1)
	//}
	//<-exit

}

// FIXME: need to figure out if this is the correct approach to gracefully
// terminate a server.
func (self *Server) Stop() {
	self.listener.Close()
	/*
		for _, lt := range self.listeners {
			lt.Close()
		}
	*/

}
