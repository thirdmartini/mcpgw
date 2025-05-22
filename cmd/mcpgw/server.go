package main

import (
	"context"
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
	"github.com/thirdmartini/mcpgw/pkg/transcriber"
	"github.com/thirdmartini/mcpgw/server"
)

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
	provider, err := createProvider(ctx, modelFlag, systemPromptFile)
	if err != nil {
		return err
	}

	host := mcphost.NewHost(provider)
	host.WithServerConfig(configFile)

	srv := server.NewServer(host)

	if whisperAddress != "" {
		log.Infof("Using Whisper transcriber at %s", whisperAddress)
		srv.WithTranscriber(transcriber.NewWhisper(whisperAddress))
	}

	return srv.ListenAndServe(serverAddress, serverRoot)
}
