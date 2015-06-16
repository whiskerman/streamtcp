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
		conn:       conn,
		incoming:   make(Message),
		outgoing:   make(Message),
		quiting:    make(chan net.Conn),
		reader:     reader,
		writer:     writer,
		messageRec: callback,
	}
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
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := self.reader.Read(buffer)
		//self.reader.Read()
		if err != nil {
			/*
				if err == io.EOF {
					//fmt.Println("n is =====================", n)
					break
				}
			*/
			log.Println(self.conn.RemoteAddr().String(), " connection error: ", err)
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
			log.Println(self.conn.RemoteAddr().String(), " recv : P ")
		} else if n > 0 {
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

func (self *Session) Write() {

	for {
		timeout := make(chan bool)
		go func() {
			time.Sleep(30 * time.Second)
			fmt.Println("sleep 30")
			timeout <- true
			fmt.Println("end sleep 30")
		}()
		select {

		case data := <-self.outgoing:
			var out []byte
			if len(data) == 1 && string(data[:1]) == "P" {
				out = data
			} else {

				out = Packet([]byte(data))
			}
			fmt.Println(self.conn, " send:", string(out))
			if _, err := self.writer.Write(out); err != nil {
				self.quit()
				return
			}
			if err := self.writer.Flush(); err != nil {
				log.Printf("Write error: %s\n", err)
				self.quit()
				return
			}
		case <-timeout:
			fmt.Println("recv sleep 30")
			go func() {
				self.outgoing <- []byte("P")
			}()
			fmt.Println("send outgoing P")
			/*
				if err := self.WritePing(); err != nil {
					self.quit()
					return
				}
			*/
			log.Println(self.conn.RemoteAddr().String(), " send : P ")
		}
	}

}

func (self *Session) Close() {
	self.conn.Close()
}
