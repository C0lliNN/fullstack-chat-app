package server

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Config struct {
	Processor *chat.ChatProcessor
	Upgrader  websocket.Upgrader
	Port int
}

type Server struct {
	Config
	server *http.Server
}

func NewServer(c Config) *Server {
	return &Server{Config: c}
}

func (s *Server) Start() {
	http.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodPost {
			chat, err := s.Processor.NewChat(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}

			data, _ := json.Marshal(chat)
			w.Header().Add("Content-Type", "application/json")
			w.Write(data)
		} else if r.Method == http.MethodGet {
			query := r.URL.Query()
			chatCode := query.Get("code")
			userName := query.Get("user")

			conn, err := s.Upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}

			err = s.Processor.JoinChat(r.Context(), chat.JoinChatRequest{
				ChatCode:   chatCode,
				UserName:   userName,
				Connection: conn,
			})

			if err != nil {
				conn.Close()
				log.Println(err)
			}
		}
	})

	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", s.Port),
	}

	log.Printf("Server starting on port %d", s.Port)
	
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server")
	return s.server.Shutdown(ctx)
}
