package main

import (
	"github.com/thirdmartini/mcpgw/pkg/mcphost"
)

type InferenceProvider struct {
	Provider     string
	Host         string
	Token        string
	Model        string
	SystemPrompt string
	Options      map[string]interface{}
}

type Config struct {
	UI struct {
		Listen string
		TLS    bool
		Root   string
	}
	SpeechToText *InferenceProvider
	TextToSpeech *InferenceProvider
	Inference    *InferenceProvider
	Servers      *mcphost.MCPConfig
}
