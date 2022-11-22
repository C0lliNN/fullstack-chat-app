package server

import (
	"c0llinn/fullstack-chat-app/internal/chat"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	processor chat.ChatProcessor
	upgrader  websocket.Upgrader
}

func NewServer(processor chat.ChatProcessor, upgrader websocket.Upgrader) *Server {
	return &Server{processor: processor, upgrader: upgrader}
}

func (s *Server) Start() {
	http.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			chat, err := s.processor.NewChat(r.Context())
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

			conn, err := s.upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}

			err = s.processor.JoinChat(r.Context(), chat.JoinChatRequest{
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

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
