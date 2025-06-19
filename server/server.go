package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/mux"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/mcphost"
	"github.com/thirdmartini/mcpgw/pkg/speaker"
	"github.com/thirdmartini/mcpgw/pkg/transcriber"
	"github.com/thirdmartini/mcpgw/server/autocert"
)

type Server struct {
	host        *mcphost.Host
	transcriber transcriber.Transcriber
	speaker     speaker.Engine

	lock     sync.Mutex
	sessions map[string]*Session
}

type Request struct {
	ConversationID string
	Prompt         string
}

type Response struct {
	ConversationID string
	Prompt         string
	Message        string
	Audio          string
	Images         []string
}

func listenStringToAddress(listen string, tls bool) string {
	var address string

	if tls {
		address = "https://"
	} else {
		address = "http://"
	}

	if strings.HasPrefix(listen, ":") {
		return address + "localhost" + listen
	}
	return address + listen
}

func (s *Server) getSession(r *http.Request) *Session {
	token := r.Header.Get("X-Conversation-Id")
	if token == "" {
		return &Session{
			id:       token,
			messages: []history.HistoryMessage{},
			window:   16,
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if session, ok := s.sessions[token]; ok {
		return session
	}

	return &Session{
		id:       token,
		messages: []history.HistoryMessage{},
		window:   16,
	}
}

func (s *Server) putSession(session *Session) {
	if session.id == "" {
		return
	}

	session.Prune()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sessions[session.id] = session
}

func (s *Server) chatErrorResponse(w http.ResponseWriter, prompt string, err error) {
	response := Response{
		Prompt:  prompt,
		Message: err.Error(),
	}
	json.NewEncoder(w).Encode(response)
}

// handleChatRequest processes a chat prompt and generates a response, optionally including audio, using the server's resources.
func (s *Server) handleChatRequest(w http.ResponseWriter, session *Session, prompt string) {
	log.Info("Chat Request", "session", session.id, "prompt", prompt)

	message, err := s.host.RunPrompt(context.Background(), prompt, &session.messages)
	if err != nil {
		log.Errorf("Error running prompt: %v", err)
		s.chatErrorResponse(w, prompt, err)
		return
	}

	response := Response{
		Prompt:  prompt,
		Message: message.Message,
		Images:  message.Images,
	}

	// if we have a speaker, convert the message to audio
	if s.speaker != nil {
		if audio, err := s.speaker.Say(message.Message); err == nil {
			data, _ := io.ReadAll(audio)
			response.Audio = base64.StdEncoding.EncodeToString(data)
			// debug
			//os.WriteFile("example/ui/static/audio.wav", data, 0644)
			//speaker.PlayAudio(response.Audio)
		}
	}
	json.NewEncoder(w).Encode(response)
}

// AudioChatRequest handles HTTP POST requests for audio-based chat interactions.
// It transcribes audio input, processes the prompt with the server's LLM host, and responds with the generated output.
func (s *Server) AudioChatRequest(w http.ResponseWriter, r *http.Request) {
	session := s.getSession(r)
	defer s.putSession(session)

	prompt, err := s.Transcribe(r.Body)
	if err != nil {
		log.Errorf("Error transcribing: %v", err)
		s.chatErrorResponse(w, "[no audio]", err)
		return
	}
	s.handleChatRequest(w, session, prompt)
}

// AudioTranscribeRequest handles HTTP POST requests for audio transcription.
// It transcribes audio input from the request body and responds with the transcribed text in JSON format.
func (s *Server) AudioTranscribeRequest(w http.ResponseWriter, r *http.Request) {
	session := s.getSession(r)
	defer s.putSession(session)

	log.Info("Audio Transcribe Request", "session", session.id)

	prompt, err := s.Transcribe(r.Body)
	if err != nil {
		log.Errorf("Error transcribing: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := Response{
		Prompt: prompt,
	}
	json.NewEncoder(w).Encode(response)
}

// ChatRequest handles HTTP POST requests for text-based chat interactions.
// It processes the request body, executes the chat prompt, and sends a JSON response with the result.
func (s *Server) ChatRequest(w http.ResponseWriter, r *http.Request) {
	session := s.getSession(r)
	defer s.putSession(session)

	log.Info("Text Chat Request", "session", session.id)
	request := Request{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.handleChatRequest(w, session, request.Prompt)
}

// GetAvailableTools handles HTTP GET requests and retrieves a list of tools available from the server's host.
// The list is returned as a JSON-encoded response.
func (s *Server) GetAvailableTools(w http.ResponseWriter, r *http.Request) {
	tools := s.host.ListTools()
	json.NewEncoder(w).Encode(tools)
}

func (s *Server) createRoutes(root string) *mux.Router {
	fs := http.Dir(root)
	router := mux.NewRouter()
	router.PathPrefix("/static").Handler(http.FileServer(fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, root+"/index.html")
	})

	router.HandleFunc("/api/v.1/tools", s.GetAvailableTools).Methods("GET")
	router.HandleFunc("/api/v.1/chat", s.ChatRequest).Methods("POST")
	router.HandleFunc("/api/v.1/recordings/save", s.AudioChatRequest).Methods("POST")
	router.HandleFunc("/api/v.1/recordings/transcribe", s.AudioTranscribeRequest).Methods("POST")

	return router
}

// ListenAndServe runs the gateway server
func (s *Server) ListenAndServe(address string, root string) error {
	srv := &http.Server{
		Handler:      s.createRoutes(root),
		Addr:         address,
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	log.Infof("Starting server at [ %s ]", listenStringToAddress(address, false))
	return srv.ListenAndServe()
}

// ListenAndServeTLS Runs a TLS version of the server
func (s *Server) ListenAndServeTLS(address string, root string, cert, key string) error {
	srv := &http.Server{
		Handler:      s.createRoutes(root),
		Addr:         address,
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
			MinVersion:               tls.VersionTLS11,
			MaxVersion:               tls.VersionTLS13,
		},
	}

	kpr, err := autocert.NewManager(cert, key)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	srv.TLSConfig.GetCertificate = kpr.GetCertificateFunc()

	log.Infof("Starting server at [ %s ]", listenStringToAddress(address, true))
	return srv.ListenAndServeTLS("", "")
}

// WithTranscriber sets the transcriber for the server and returns the updated server instance.
// a transcriber will convert spoken audio to text, we include a sample transcriber that uses whisper-server
func (s *Server) WithTranscriber(transcriber transcriber.Transcriber) *Server {
	s.transcriber = transcriber
	return s
}

// WithAudioEncoder sets the audio speaker engine for the server and returns the updated server instance.
// the speaker engine will convert the llm return into audio and send it back the audio as part of the llm response
// we include an example that uses the MeloTTS engine, feel free to add your own
func (s *Server) WithAudioEncoder(speaker speaker.Engine) *Server {
	s.speaker = speaker
	return s
}

func NewServer(host *mcphost.Host) *Server {
	log.SetLevel(log.DebugLevel)

	return &Server{
		host:     host,
		sessions: make(map[string]*Session),
	}
}
