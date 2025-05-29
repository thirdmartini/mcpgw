package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/thirdmartini/mcpgw/pkg/llm"
	"github.com/thirdmartini/mcpgw/pkg/llm/anthropic"
	"github.com/thirdmartini/mcpgw/pkg/llm/google"
	"github.com/thirdmartini/mcpgw/pkg/llm/ollama"
	"github.com/thirdmartini/mcpgw/pkg/llm/openai"
	"github.com/thirdmartini/mcpgw/pkg/mcphost"
	"github.com/thirdmartini/mcpgw/pkg/speaker"
	"github.com/thirdmartini/mcpgw/pkg/transcriber"
	"github.com/thirdmartini/mcpgw/server"
)

// loadSystemPrompt loads the system prompt from a JSON file
func loadSystemPrompt(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading config file: %v", err)
	}

	// Parse only the systemPrompt field
	var config struct {
		SystemPrompt string `json:"systemPrompt"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("error parsing config file: %v", err)
	}

	return config.SystemPrompt, nil
}

func createProvider(ctx context.Context, modelString, systemPrompt string) (llm.Provider, error) {
	parts := strings.SplitN(modelString, ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf(
			"invalid model format. Expected provider:model, got %s",
			modelString,
		)
	}

	provider := parts[0]
	model := parts[1]

	switch provider {
	case "anthropic":
		apiKey := anthropicAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}

		if apiKey == "" {
			return nil, fmt.Errorf(
				"Anthropic API key not provided. Use --anthropic-api-key flag or ANTHROPIC_API_KEY environment variable",
			)
		}
		return anthropic.NewProvider(apiKey, anthropicBaseURL, model, systemPrompt), nil

	case "ollama":
		return ollama.NewProvider(model, systemPrompt)

	case "openai":
		apiKey := openaiAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}

		if apiKey == "" {
			return nil, fmt.Errorf(
				"OpenAI API key not provided. Use --openai-api-key flag or OPENAI_API_KEY environment variable",
			)
		}
		return openai.NewProvider(apiKey, openaiBaseURL, model, systemPrompt), nil

	case "google":
		apiKey := googleAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			// The project structure is provider specific, but Google calls this GEMINI_API_KEY in e.g. AI Studio. Support both.
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		return google.NewProvider(ctx, apiKey, model, systemPrompt)

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func runServer(ctx context.Context) error {
	systemPrompt, _ := loadSystemPrompt(systemPromptFile)

	log.Infof("SystemPrompt: %s\n", systemPrompt)

	provider, err := createProvider(ctx, modelFlag, systemPrompt)
	if err != nil {
		return err
	}

	host := mcphost.NewHost(provider)
	host.WithServerConfig(configFile)

	srv := server.NewServer(host)

	if whisperAddress != "" {
		log.Infof("Using Whisper for Speech To Text at %s", whisperAddress)
		srv.WithTranscriber(transcriber.NewWhisper(whisperAddress))
	}

	if meloAddress != "" {
		log.Infof("Using MeloTTS for Text To Speech at %s", meloAddress)
		srv.WithAudioEncoder(speaker.NewMelo(speaker.MeloOptions{
			Address: meloAddress,
			Voice:   "",
		}))
	}

	if strings.HasPrefix(serverAddress, ":") {
		log.Infof("Starting server at %s | http://localhost%s", serverAddress, serverAddress)
	} else {
		log.Infof("Starting server at %s | http://%s", serverAddress, serverAddress)
	}

	tls := true
	if tls {
		return srv.ListenAndServeTLS(serverAddress, serverRoot, "cert.pem", "key.pem")
	} else {
		return srv.ListenAndServe(serverAddress, serverRoot)
	}

}
