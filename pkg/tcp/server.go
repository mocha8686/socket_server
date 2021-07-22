package tcp

import (
	"strings"
)

// handler interface
type handler interface {
	OnConnect(*Client)
	OnMessage(*Client, string)
	OnClose(string)
}

// Helper functions
func parseMessage(raw string) (directive, content string) {
	final := strings.SplitN(raw, ":", 2)
	return final[0], final[1]
}


// Server
type server struct {
	listener Listener
	handler handler
	clientChan <-chan *Client
	msgChan <-chan string

	Clients map[string]*Client
}

// Constructor
func NewServer(handler handler) server {
	clientChan := make(chan *Client)
	msgChan := make(chan string)

	var readClientChan <-chan *Client = clientChan
	var readMsgChan <-chan string = msgChan

	var writeClientChan chan<- *Client = clientChan
	var writeMsgChan chan<- string = msgChan

	return server{
		listener: NewListener(writeClientChan, writeMsgChan),
		handler: handler,
		clientChan: readClientChan,
		msgChan: readMsgChan,

		Clients: make(map[string]*Client, 0),
	}
}

// Methods
func (s *server) Start(port int) {
	go s.listener.Start(port)

	for {
		select {
		case newClient := <-s.clientChan:
			go s.handler.OnConnect(newClient)
			go newClient.Receive()
			s.Clients[newClient.ID] = newClient

		case raw := <-s.msgChan:
			dir, content := parseMessage(raw)

			if dir == "CLOSE" {
				if s.Clients[content] != nil {
					delete(s.Clients, content)
					go s.handler.OnClose(content)
					break
				}
			}

			go s.handler.OnMessage(s.Clients[dir], content)
		}
	}
}
