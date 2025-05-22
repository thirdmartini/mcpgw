package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/mux"

	"github.com/thirdmartini/mcpgw/pkg/mcphost"
	"github.com/thirdmartini/mcpgw/pkg/transcriber"
)

type Server struct {
	host        *mcphost.Host
	transcriber transcriber.Transcriber
}

type Request struct {
	Prompt string
}

type Response struct {
	Message string
}

type Session struct {
	//
}

func (s *Server) AudioChatRequest(w http.ResponseWriter, r *http.Request) {
	log.Info("Audio Chat Request")

	prompt, err := s.Transcribe(r.Body)
	if err != nil {
		log.Errorf("Error transcribing: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("Transcribed Prompt: " + prompt)

	message, err := s.host.RunPrompt(context.Background(), prompt)
	log.Info("LLM Response: " + message)
	json.NewEncoder(w).Encode(Response{Message: message})
}

func (s *Server) ChatRequest(w http.ResponseWriter, r *http.Request) {
	request := Request{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message, err := s.host.RunPrompt(context.Background(), request.Prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(Response{Message: message})
}

func (s *Server) ListenAndServe(address string, root string) error {
	router := mux.NewRouter()

	srv := &http.Server{
		Handler: router,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	fs := http.Dir(root)
	router.PathPrefix("/static").Handler(http.FileServer(fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, root+"/index.html")
	})
	router.HandleFunc("/api/v.1/chat", s.ChatRequest).Methods("POST")
	router.HandleFunc("/api/v.1/recordings/save", s.AudioChatRequest).Methods("POST")

	return srv.ListenAndServe()
}

func (s *Server) WithTranscriber(t transcriber.Transcriber) *Server {
	s.transcriber = t
	return s
}

func NewServer(host *mcphost.Host) *Server {
	return &Server{
		host: host,
	}
}
