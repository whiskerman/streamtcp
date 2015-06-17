package streamtcp

import (
	"bufio"
	"errors"
	//"fmt"
	"log"
	"net"
)

type CallBackCClient func(*Client, string)

type Client struct {
	conn       net.Conn
	incoming   Message
	outgoing   Message
	reader     *bufio.Reader
	writer     *bufio.Writer
	quiting    chan net.Conn
	name       string
	messageRec CallBackCClient
}

func (self *Client) GetName() string {
	return self.name
}

func (self *Client) GetConn() net.Conn {
	return self.conn
}

func (self *Client) SetName(name string) {
	self.name = name
}

func (self *Client) GetIncoming() []byte {
	return <-self.incoming
}

func (self *Client) PutOutgoing(message []byte) {
	self.outgoing <- message
}

func CreateClient(conn net.Conn, callback CallBackCClient) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	Client := &Client{
		conn:       conn,
		incoming:   make(Message, 1024),
		outgoing:   make(Message, 1024),
		quiting:    make(chan net.Conn),
		reader:     reader,
		writer:     writer,
		messageRec: callback,
	}
	Client.Listen()
	return Client
}

func (self *Client) Listen() {
	go self.Read()
	go self.Write()
}

func (self *Client) quit() {
	self.quiting <- self.conn
}

func (self *Client) WritePing() error {
	if _, err := self.writer.Write([]byte("P")); err != nil {

		return errors.New("write ping error")
	}
	if err := self.writer.Flush(); err != nil {
		log.Printf("Write error: %s\n", err)

		return errors.New("write ping error")
	}
	return nil
}

func (self *Client) Read() {
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

func (self *Client) Write() {
	for {
		datas, _ := <-self.outgoing
		out := Packet([]byte(datas))
		//fmt.Println(string(out))
		if _, err := self.writer.Write(out); err != nil {
			log.Printf("Write error: %s\n", err)
			self.quit()
			return
		}
		if err := self.writer.Flush(); err != nil {
			log.Printf("Write error: %s\n", err)
			self.quit()
			return
		}

	}

}

func (self *Client) Close() {
	self.conn.Close()
}
