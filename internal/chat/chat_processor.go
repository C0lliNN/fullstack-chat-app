package chat

import (
	"context"
	"log"
	"time"
)

type ChatRepository interface {
	Save(ctx context.Context, chat Chat) error
	FindByCode(ctx context.Context, code string) (Chat, error)
}

type MessageRepository interface {
	Save(ctx context.Context, message Message) error
}

type CodeGenerator interface {
	NewCode() string
}

type IDGenerator interface {
	NewID() string
}

type MessageMarshaller interface {
	Marshal(data interface{}) ([]byte, error)
	Unmarshal(data interface{}) ([]byte, error)
}

type ProcessorConfig struct {
	MessageRepository MessageRepository
	ChatRepository    ChatRepository
	CodeGenerator     CodeGenerator
	IDGenerator       IDGenerator
}

type ChatProcessor struct {
	ProcessorConfig
	chats map[*ChatRoom]bool
}

func NewChatProcessor(c ProcessorConfig) *ChatProcessor {
	processor := &ChatProcessor{ProcessorConfig: c, chats: make(map[*ChatRoom]bool)}
	go processor.cleanEmptyChats()

	return processor
}

func (p *ChatProcessor) NewChat(ctx context.Context) (Chat, error) {
	chat := Chat{ID: p.IDGenerator.NewID(), Code: p.CodeGenerator.NewCode()}
	if err := p.ChatRepository.Save(ctx, chat); err != nil {
		return Chat{}, err
	}

	p.newChatRoom(chat)

	log.Println("New chat room was created successfully")

	return chat, nil
}

type JoinChatRequest struct {
	ChatCode   string
	UserName   string
	Connection Connection
}

func (p *ChatProcessor) JoinChat(ctx context.Context, req JoinChatRequest) error {
	chat, err := p.ChatRepository.FindByCode(ctx, req.ChatCode)
	if err != nil {
		return err
	}

	var chatRoom *ChatRoom
	if !p.isChatCached(chat) {
		chatRoom = p.newChatRoom(chat)
	} else {
		chatRoom = p.findCachedChatRoom(chat)
	}

	chatRoom.HandleUserConnected(context.Background(), req.UserName, req.Connection)
	log.Printf("New user joined the chat %s", chat.Code)

	return nil
}

func (p *ChatProcessor) isChatCached(chat Chat) bool {
	for c := range p.chats {
		if c.Chat.ID == chat.ID {
			return true
		}
	}

	return false
}

func (p *ChatProcessor) findCachedChatRoom(chat Chat) *ChatRoom {
	for c := range p.chats {
		if c.Chat.ID == chat.ID {
			return c
		}
	}

	return nil
}

func (p *ChatProcessor) newChatRoom(chat Chat) *ChatRoom {
	chatRoom := NewChatRoom(RoomConfig{
		IDGenerator: p.IDGenerator,
		Repository:  p.MessageRepository,
		Chat:        chat,
	})

	p.chats[chatRoom] = true

	return chatRoom
}

func (p *ChatProcessor) cleanEmptyChats() {
	for range time.Tick(time.Minute) {
		for chat := range p.chats {
			if chat.Empty() {
				log.Printf("Chat %s is empty, so removing it from the cache", chat.Chat.Code)
				delete(p.chats, chat)
			}
		}
	}
}
