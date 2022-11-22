package persistence

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct{
	collection *mongo.Collection
}

func NewMessageRepository(collection *mongo.Collection) MessageRepository {
	return MessageRepository{collection: collection}
}

func (r MessageRepository) Save(ctx context.Context, message chat.Message) error {
	_, err := r.collection.InsertOne(ctx, message)
	return err
}