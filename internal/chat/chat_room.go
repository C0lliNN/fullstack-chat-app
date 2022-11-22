package chat

import (
	"context"
	"encoding/json"
	"log"
)

type MessageChannel chan []byte

type RoomConfig struct {
	Repository       MessageRepository
	IDGenerator      IDGenerator
	Chat             Chat
	BroadcastChannel MessageChannel
}

type ChatRoom struct {
	RoomConfig
	clients map[*Client]bool
}

func NewChatRoom(c RoomConfig) *ChatRoom {
	return &ChatRoom{RoomConfig: c, clients: make(map[*Client]bool)}
}

type InsertMessageRequest struct {
	Content string
	User    User
}

func (r *ChatRoom) HandleNewMessage(ctx context.Context, req InsertMessageRequest) error {
	m := Message{
		ID:      r.IDGenerator.NewID(),
		Content: req.Content,
		ChatID:  r.Chat.ID,
		User:    req.User,
	}

	if err := r.Repository.Save(ctx, m); err != nil {
		return err
	}
	log.Println("New message saved successfully")

	rawMessage, err := json.Marshal(m)
	if err != nil {
		return err
	}

	for client := range r.clients {
		client.ClientChannel <- rawMessage
	}

	log.Println("New message processed successfully")

	return nil
}

func (r *ChatRoom) HandleUserConnected(ctx context.Context, userName string, connection Connection) {
	client := NewChatClient(ClientConfig{
		ClientChannel:    make(MessageChannel),
		Connection:       connection,
		OnMessageHandler: r.HandleNewMessage,
		User:             User{ID: r.IDGenerator.NewID(), Name: userName},
	})

	r.clients[client] = true

	go client.ListenForClientChannelWrites(ctx)
	go client.ListenForConnectionWrites(ctx)

	log.Println("New user connected")
}

func (r *ChatRoom) HandleUserDisconnected(ctx context.Context, userId string) {
	for client := range r.clients {
		if client.User.ID == userId {
			client.Connection.Close()
			delete(r.clients, client)
			ctx.Done()
			log.Println("Client disconnected")
			break
		}
	}
}

func (r *ChatRoom) Empty() bool {
	return len(r.clients) == 0
}
