package speaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var emojiRx = regexp.MustCompile(`[\x{1F600}-\x{1F6FF}|[\x{2600}-\x{26FF}]|[\x{1F300}-\x{1F5FF}]`)

type MeloOptions struct {
	Address string
	Voice   string
}

type Melo struct {
	opts MeloOptions
	//	Address string
	//	Voice   string
}

type MeloRequest struct {
	Voice string `json:"voice"`
	Text  string `json:"text"`
}

func (s *Melo) Say(text string) (SpeechStream, error) {
	mr := MeloRequest{
		Voice: s.opts.Voice,
		Text:  emojiRx.ReplaceAllString(text, ``),
	}

	data, err := json.Marshal(&mr)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/v1/inference", s.opts.Address), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	defer resp.Body.Close()
	return nil, io.EOF
}

func NewMelo(opts MeloOptions) *Melo {
	return &Melo{
		opts: opts,
	}
}
