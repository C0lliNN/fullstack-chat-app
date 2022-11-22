package persistence

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(collection *mongo.Collection) ChatRepository {
	return ChatRepository{collection: collection}
}

func (r ChatRepository) Save(ctx context.Context, chat chat.Chat) error {
	_, err := r.collection.InsertOne(ctx, chat)
	return err
}

func (r ChatRepository) FindByCode(ctx context.Context, code string) (chat.Chat, error) {
	var c chat.Chat
	
	result := r.collection.FindOne(ctx, bson.M{"code": code})
	if err := result.Decode(&c); err != nil {
		return chat.Chat{}, err
	}

	return c, nil
}