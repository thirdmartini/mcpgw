package transcriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type WhisperClient struct {
	Address string
}

type Response struct {
	Text string `json:"text"`
}

func (c *WhisperClient) Transcribe(r io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "audio")
	io.Copy(part, r)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/inference", c.Address), body)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		response := Response{}

		if err = json.Unmarshal(content, &response); err != nil {
			return "", nil
		}

		return response.Text, nil
	}
	return "", fmt.Errorf("%s", resp.Status)
}

func (c *WhisperClient) TranscribeFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return c.Transcribe(f)
}

func NewWhisper(address string) *WhisperClient {
	return &WhisperClient{
		Address: address,
	}
}
