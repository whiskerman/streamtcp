package streamtcp

import (
	"fmt"
	"log"
	"net"
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
		sessions: make(ClientTable, MAXCLIENTS),
		tokens:   make(Token, MAXCLIENTS),
		pending:  make(chan net.Conn, 1024),
		quiting:  make(chan net.Conn, 1024),
		incoming: make(Message, 10240),
		outgoing: make(Message, 10240),

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
				self.broadcast(message)
			case conn := <-self.pending:
				self.join(conn)
			case conn := <-self.quiting:
				self.leave(conn)
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
	if conn != nil {
		conn.Close()
		delete(self.sessions, conn)
	}

	self.generateToken()
}

func (self *Server) broadcast(message []byte) {
	//log.Printf("Broadcasting message: %s\n", message)
	go func() {
		for _, client := range self.sessions {
			func(se *Session) {
				if !se.closing {
					se.outgoing <- message
				}
			}(client)
			//fmt.Println(client.conn, ":", message)
		}
	}()
}

func (self *Server) Start(connString string) {

	func(n int, constr string) {
		var err error
		self.listener, err = net.Listen("tcp", constr)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Server %p starts%v\n  x is %d", self, self, n)

		// filling the tokens
		for i := 0; i < MAXCLIENTS; i++ {
			self.generateToken()
		}

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

	}(1, connString)

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
