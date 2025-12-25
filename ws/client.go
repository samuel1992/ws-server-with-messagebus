package ws

import (
	"log"
	"sync"
	"time"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn         *websocket.Conn
	messageBus   messagebus.MessageBus
	readTopic    string
	writeTopic   string
	sendToWsConn chan []byte
	done         chan struct{}
}

func NewClient(conn *websocket.Conn, mb messagebus.MessageBus, readTopic, writeTopic string) *Client {
	return &Client{
		conn:       conn,
		messageBus: mb,
		readTopic:  readTopic,
		writeTopic: writeTopic,
		done:       make(chan struct{}),
	}
}

// readLoop reads messages from the websocket and writes them to the message bus.
func (c *Client) readLoop() {
	defer func() {
		c.messageBus.Unsubscribe(c.readTopic, c.sendToWsConn)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.messageBus.Publish(c.writeTopic, message)
	}
}

// writeLoop writes messages from the message bus to the websocket connection.
func (c *Client) writeLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.sendToWsConn:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Start() error {
	c.sendToWsConn = c.messageBus.Subscribe(c.readTopic)

	var wg sync.WaitGroup

	wg.Go(c.writeLoop)
	wg.Go(c.readLoop)

	go func() {
		wg.Wait()
		close(c.done)
	}()

	<-c.done

	return nil
}

func (c *Client) Stop() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
