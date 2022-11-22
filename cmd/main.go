package main

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"c0llinn/fullstack-chat-app/internal/generator"
	"c0llinn/fullstack-chat-app/internal/marshaler"
	"c0llinn/fullstack-chat-app/internal/persistence"
	"c0llinn/fullstack-chat-app/internal/server"
	"context"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		panic(err)
	}

	database := client.Database(os.Getenv("MONGO_DATABASE"))

	chatRepo := persistence.NewChatRepository(database.Collection("chats"))
	messageRepo := persistence.NewMessageRepository(database.Collection("messages"))

	idGenerator := generator.NewUUIDGenerator()
	codeGenerator := generator.NewCodeGenerator()

	marshaller := marshaler.NewJSONMarshaller()

	processor := chat.NewChatProcessor(chat.ProcessorConfig{
		IDGenerator:       idGenerator,
		CodeGenerator:     codeGenerator,
		ChatRepository:    chatRepo,
		MessageRepository: messageRepo,
		Marshaller:        marshaller,
	})

	server := server.NewServer(*processor, websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	server.Start()
}
