package main

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"c0llinn/fullstack-chat-app/internal/generator"
	"c0llinn/fullstack-chat-app/internal/marshaler"
	"c0llinn/fullstack-chat-app/internal/persistence"
	"c0llinn/fullstack-chat-app/internal/server"
	"context"
	"log"
	"os"
	"os/signal"
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

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	server := server.NewServer(server.Config{Processor: processor, Upgrader: upgrader, Port: 3000})

	shutdown := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Println("Cleaning up resources")
		
		ctx, cancel = context.WithTimeout(context.Background(), time.Second * 20)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}

		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Mongo Disconnect Error: %v", err)
		}

		processor.Shutdown()

		close(shutdown)
	}()

	server.Start()

	<-shutdown
}
