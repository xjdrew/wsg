package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	// The websocket connection from client
	clientConn *websocket.Conn

	// connection to server
	serverConn *ServerConn
}

// readPump pumps messages from the websocket connection to the server.
func (c *Client) readPump() {
	defer func() {
		c.clientConn.Close()
		c.serverConn.Close()
	}()
	c.clientConn.SetReadLimit(maxMessageSize)
	c.clientConn.SetReadDeadline(time.Now().Add(pongWait))
	c.clientConn.SetPongHandler(func(string) error { c.clientConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.clientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if _, err = c.serverConn.Write(message); err != nil {
			log.Printf("error: %v", err)
			break
		}
	}
}

// writePump pumps messages from the server to the websocket connection.
func (c *Client) writePump() {
	defer func() {
		c.serverConn.Close()
		c.clientConn.Close()
	}()
	b := make([]byte, maxMessageSize)
	for {
		n, err := c.serverConn.Read(b)
		if err != nil {
			// The hub closed the channel.
			c.clientConn.WriteMessage(websocket.CloseMessage, []byte{})
			log.Printf("error: %v", err)
			break
		}

		c.clientConn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.clientConn.WriteMessage(websocket.TextMessage, b[:n]); err != nil {
			log.Printf("error: %v", err)
			break
		}
	}
}

func (c *Client) run() {
	go c.readPump()
	go c.writePump()
}

// serveWs handles websocket requests from the peer.
func serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	serverConn, err := server.Dial()
	if err != nil {
		clientConn.Close()
		log.Println(err)
		return
	}
	client := &Client{clientConn: clientConn, serverConn: serverConn}
	client.run()
}
