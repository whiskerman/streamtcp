package streamtcp

import (
	"bufio"
	//	"errors"
	"fmt"

	"log"
	"net"
	"time"
)

type CallBackClient func(*Session, string)

type Session struct {
	conn       net.Conn
	incoming   Message
	outgoing   Message
	reader     *bufio.Reader
	writer     *bufio.Writer
	quiting    chan net.Conn
	name       string
	closing    bool
	messageRec CallBackClient
}

func (self *Session) GetName() string {
	return self.name
}

func (self *Session) SetName(name string) {
	self.name = name
}

func (self *Session) GetIncoming() []byte {
	return <-self.incoming
}

func (self *Session) PutOutgoing(message []byte) {
	self.outgoing <- message
}

func CreateSession(conn net.Conn, callback CallBackClient) *Session {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	session := &Session{
		conn:     conn,
		incoming: make(Message, 1024),
		outgoing: make(Message, 1024),
		quiting:  make(chan net.Conn),
		reader:   reader,
		writer:   writer,

		messageRec: callback,
	}
	session.closing = false
	session.Listen()
	return session
}

func (self *Session) Listen() {
	go self.Read()
	go self.Write()
}

func (self *Session) quit() {
	self.quiting <- self.conn

}

/*
func (self *Session) WritePing() error {
	if _, err := self.writer.Write([]byte("P")); err != nil {

		return errors.New("write ping error")
	}
	if err := self.writer.Flush(); err != nil {
		log.Printf("Write error: %s\n", err)

		return errors.New("write ping error")
	}
	return nil
}
*/

func (self *Session) Read() {
	if self.closing {
		return
	}
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		if self.closing {
			return
		}
		n, err := self.reader.Read(buffer)
		//self.reader.Read()
		if err != nil {
			/*
				if err == io.EOF {
					//fmt.Println("n is =====================", n)
					break
				}
			*/
			self.closing = true
			log.Println(" connection error: ", err) //self.conn.RemoteAddr().String(),
			self.quit()
			return
		}
		if n == 1 && string(buffer[:1]) == "P" {
			/*
				if _, err := self.writer.Write([]byte("P")); err != nil {
					self.quit()
					return
				}
				if err := self.writer.Flush(); err != nil {
					log.Printf("Write error: %s\n", err)
					self.quit()
					return
				}
			*/
			//log.Println(self.conn.RemoteAddr().String(), " recv : P ")
		}
		if n > 0 {
			//fmt.Println("n is ========================================", n)
			tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), self.incoming)

		}
		/*
			if line, _, err := self.reader.ReadLine(); err == nil {
				self.incoming <- string(line)
			} else {
				log.Printf("Read error: %s\n", err)
				self.quit()
				return
			}
		*/
	}

}
func (self *Session) WritePing() {
	self.outgoing <- []byte("P")
}
func (self *Session) Write() {

	for {
		if self.closing {
			return
		}
		/*
			timeout := make(chan bool)
			defer func() {
				//close(timeout)
				//<-timeout
			}()
			go func() {
				if self.closing {
					close(timeout)
					return
				}
				time.Sleep(30 * time.Second)
				//fmt.Println("sleep 30")
				if self.closing {
					close(timeout)
					return
				}
				timeout <- true
				//fmt.Println("end sleep 30")
			}()
		*/
		select {

		case <-time.After(time.Second * 30):
			if self.closing {
				return
			}
			//close(timeout)
			//fmt.Println("my time out")
			//fmt.Println("recv sleep 30")
			go self.WritePing()
			//fmt.Println("send outgoing P")
			/*
				if err := self.WritePing(); err != nil {
					self.quit()
					return
				}
			*/
			//log.Println(self.conn.RemoteAddr().String(), " send : P ")

		case data := <-self.outgoing:
			if self.closing {
				return
			}
			var out []byte
			if len(data) == 1 && string(data[:1]) == "P" {
				out = data
				fmt.Println("my time out")
			} else {

				out = Packet([]byte(data))
			}
			//fmt.Println(self.conn, " send:", string(out))
			if _, err := self.writer.Write(out); err != nil {
				log.Printf("Write error: %s\n", err)
				self.closing = true
				self.quit()
				return
			}
			if err := self.writer.Flush(); err != nil {
				log.Printf("Write error: %s\n", err)
				self.closing = true
				self.quit()
				return
			}
		}
		//case <-timeout:

	}

}

func (self *Session) Close() {
	self.conn.Close()
}
