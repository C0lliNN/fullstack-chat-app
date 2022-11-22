package chat

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	socketCloseMessage = 8
	socketTextMessage  = 1
	socketPingMessage  = 9
)

// Connection A representation of the network connection between the client and chat
type Connection interface {
	SetWriteDeadline(time.Time) error
	SetReadLimit(int64)
	SetReadDeadline(t time.Time) error
	SetPongHandler(func(string) error)
	ReadMessage() (int, []byte, error)
	NextWriter(int) (io.WriteCloser, error)
	WriteMessage(int, []byte) error
	Close() error
}

type ClientConfig struct {
	WriteWait time.Duration

	// PongWait Time allowed to read the next pong message from the peer.
	PongWait time.Duration

	// PingPeriod Send pings to peer with this period. Must be less than pongWait.
	PingPeriod time.Duration

	// MaxMessageSize Maximum message size allowed from peer.
	MaxMessageSize time.Duration

	// ClientChannel channel used to send and receive message to/from the client connection
	ClientChannel MessageChannel

	// Marshaller interface used for unmarshaling the raw message into a struct
	Marshaller MessageMarshaller

	// Connection the client connection
	Connection Connection

	// OnMessageHandler code to be processed when a new message is detected
	OnMessageHandler func(ctx context.Context, req InsertMessageRequest) error

	// OnMessageHandler code to be processed when a new message is detected
	OnDisconnectedHandler func(ctx context.Context, userId string)

	// User represents actual chat user information behind the client
	User User
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ClientConfig
}

func NewChatClient(c ClientConfig) *Client {
	if c.WriteWait == 0 {
		c.WriteWait = 10 * time.Second
	}

	if c.PongWait == 0 {
		c.PongWait = 60 * time.Second
	}

	if c.PingPeriod == 0 {
		c.PingPeriod = 45 * time.Second
	}

	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 512
	}

	return &Client{ClientConfig: c}
}

// ListenForConnectionWrites constantly reads the client Connection
// When a new message is found, send it to the Broadcast channel
func (c *Client) ListenForConnectionWrites(ctx context.Context) {
	defer func() {
		log.Println("[ListenForConnectionWrites] Terminating client connection")
		c.OnDisconnectedHandler(ctx, c.User.ID)
	}()

	c.Connection.SetReadLimit(int64(c.MaxMessageSize))
	c.Connection.SetReadDeadline(time.Now().Add(c.PongWait))
	c.Connection.SetPongHandler(func(s string) error {
		c.Connection.SetReadDeadline(time.Now().Add(c.PongWait))
		return nil
	})

	for {
		_, rawMessage, err := c.Connection.ReadMessage()
		if err != nil {
			log.Printf("[ReadMessage] error: %v", err)
			break
		}
		rawMessage = bytes.TrimSpace(bytes.Replace(rawMessage, newline, space, -1))
		log.Printf("Received Message: %s", string(rawMessage))

		var req InsertMessageRequest
		err = c.Marshaller.Unmarshal(rawMessage, &req)
		if err != nil {
			log.Printf("[Unmarshal] error: %v", err)
			break
		}

		req.User = c.User

		if err = c.OnMessageHandler(ctx, req); err != nil {
			log.Printf("[MessageHandler] error: %v", err)
			break
		}
	}
}

// ListenForClientChannelWrites constantly reads the ClientChannel
// When a new message is found, it means a peer client wrote a message to the BroadcastChannel
// In this case, it's needed to push the message to the client connection, so it can handle it
func (c *Client) ListenForClientChannelWrites(ctx context.Context) {
	ticker := time.NewTicker(c.PingPeriod)

	defer func() {
		ticker.Stop()
		log.Println("[ListenForClientChannelWrites] Terminating client connection")
		c.OnDisconnectedHandler(ctx, c.User.ID)
	}()

	for {
		select {
		case message, ok := <-c.ClientChannel:
			log.Printf("Detected a new message in the client channel for the user: %v", c.User.ID)

			c.Connection.SetWriteDeadline(time.Now().Add(c.WriteWait))
			if !ok {
				c.Connection.WriteMessage(socketCloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(socketTextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.ClientChannel)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.ClientChannel)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(c.WriteWait))
			if err := c.Connection.WriteMessage(socketPingMessage, nil); err != nil {
				return
			}
		}
	}
}
