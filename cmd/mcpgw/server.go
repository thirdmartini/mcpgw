package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

func createInferenceProvider(ctx context.Context, config *InferenceProvider) (llm.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("inference provider not provided")
	}

	switch config.Provider {
	case "anthropic":
		return anthropic.NewProvider(config.Token, config.Host, config.Model, config.SystemPrompt), nil

	case "ollama":
		return ollama.NewProvider(config.Host, config.Model, config.SystemPrompt)

	case "openai":
		return openai.NewProvider(config.Token, config.Host, config.Model, config.SystemPrompt), nil

	case "google":
		return google.NewProvider(ctx, config.Token, config.Model, config.SystemPrompt)

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

func createSpeechToTextProvider(config *InferenceProvider) (transcriber.Transcriber, error) {
	if config == nil {
		return nil, fmt.Errorf("speech to text provider not provided")
	}

	switch config.Provider {
	case "whisper":
		return transcriber.NewWhisper(config.Host), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

func createTextToSpeechProvider(config *InferenceProvider) (speaker.Engine, error) {
	if config == nil {
		return nil, fmt.Errorf("speech to text provider not provided")
	}

	switch config.Provider {
	case "melo", "melotts":
		return speaker.NewMelo(speaker.MeloOptions{
			Address: config.Host,
			Voice:   "",
		}), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

func loadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	config := &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func runServer(ctx context.Context) error {
	config, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	provider, err := createInferenceProvider(ctx, config.Inference)
	if err != nil {
		return err
	}

	host := mcphost.NewHost(provider)
	host.WithConfig(config.Servers)
	srv := server.NewServer(host)

	log.Infof("Using Inference provider: %sn", provider.Name())
	if transcriber, err := createSpeechToTextProvider(config.SpeechToText); err == nil {
		log.Infof("Using speech to text provider: %s", config.SpeechToText.Provider)
		srv.WithTranscriber(transcriber)
	}

	if speaker, err := createTextToSpeechProvider(config.TextToSpeech); err == nil {
		log.Infof("Using speech to text provider: %s", config.TextToSpeech.Provider)
		srv.WithAudioEncoder(speaker)
	}

	if config.UI.TLS {
		return srv.ListenAndServeTLS(config.UI.Listen, config.UI.Root, "cert.pem", "key.pem")
	} else {
		return srv.ListenAndServe(config.UI.Listen, config.UI.Root)
	}
}
