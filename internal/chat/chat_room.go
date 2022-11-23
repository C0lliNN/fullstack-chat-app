package chat

import (
	"context"
	"log"
	"sync"
)

type MessageChannel chan []byte

type RoomConfig struct {
	Repository       MessageRepository
	IDGenerator      IDGenerator
	Chat             Chat
	BroadcastChannel MessageChannel
	Marshaler        MessageMarshaller
}

type ChatRoom struct {
	RoomConfig
	clients map[*Client]bool
	mutex   sync.Mutex
}

func NewChatRoom(c RoomConfig) *ChatRoom {
	return &ChatRoom{RoomConfig: c, clients: make(map[*Client]bool)}
}

type InsertMessageRequest struct {
	Content string
	User    User
}

func (r *ChatRoom) HandleNewMessage(ctx context.Context, req InsertMessageRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

	rawMessage, err := r.Marshaler.Marshal(m)
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	client := NewChatClient(ClientConfig{
		ClientChannel:         make(MessageChannel),
		Connection:            connection,
		Marshaller:            r.Marshaler,
		OnMessageHandler:      r.HandleNewMessage,
		OnDisconnectedHandler: r.HandleUserDisconnected,
		User:                  User{ID: r.IDGenerator.NewID(), Name: userName},
	})

	r.clients[client] = true

	go client.ListenForClientChannelWrites(ctx)
	go client.ListenForConnectionWrites(ctx)

	log.Println("New user connected")
}

func (r *ChatRoom) HandleUserDisconnected(ctx context.Context, userId string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.clients {
		if client.User.ID == userId {
			client.Close()
			delete(r.clients, client)
			ctx.Done()
			log.Println("Client disconnected")
			break
		}
	}
}

func (r *ChatRoom) Empty() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return len(r.clients) == 0
}

func (r *ChatRoom) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.clients {
		client.Close()
	}
}
