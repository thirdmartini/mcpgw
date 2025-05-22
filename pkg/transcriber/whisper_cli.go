package transcriber

import (
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"
)

type WhisperCli struct {
	Path string
}

func (c *WhisperCli) Transcribe(r io.Reader) (string, error) {
	f, err := os.CreateTemp("/tmp", "sample")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())

	_, err = io.Copy(f, r)
	if err != nil {
		return "", err
	}

	return c.TranscribeFile(f.Name())
}

func (c *WhisperCli) TranscribeFile(name string) (string, error) {
	out, err := exec.Command("./tools/whisper",
		"-m", "./tools/models/ggml-small.en.bin",
		"--output-txt",
		"-f", name,
		"-of", name,
	).CombinedOutput()

	if err != nil {
		log.Debugf("whisper:%s\n", string(out))
		return "", err
	}

	data, err := os.ReadFile(name + ".txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewWhisperCli contains the path to the whisper cli install (including models)
func NewWhisperCli(path string) *WhisperCli {
	return &WhisperCli{
		Path: path,
	}
}
