package speaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	TTSInferencePath = "/api/v.1/inference/tts"
)

type VoiceRouterClient struct {
	address string
	token   string
	engine  string
	voice   string
}

type VoiceRouterOptions struct {
	Address string
	Voice   string
	Engine  string
	Token   string
}

type VRSpeechRequest struct {
	Engine string `json:"engine"`
	Voice  string `json:"voice"`
	Text   string `json:"text"`
}

func (s *VoiceRouterClient) Say(text string) (SpeechStream, error) {
	path := fmt.Sprintf("%s%s", s.address, TTSInferencePath)

	mr := VRSpeechRequest{
		Engine: s.engine,
		Voice:  s.voice,
		Text:   emojiRx.ReplaceAllString(text, ``),
	}

	data, err := json.Marshal(&mr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: s.token,
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	defer resp.Body.Close()
	return nil, fmt.Errorf(resp.Status)
}

func NewVoiceRouterClient(opts VoiceRouterOptions) *VoiceRouterClient {
	return &VoiceRouterClient{
		address: opts.Address,
		token:   opts.Token,
		engine:  opts.Engine,
		voice:   opts.Voice,
	}
}
