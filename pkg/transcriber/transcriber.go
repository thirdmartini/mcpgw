package transcriber

import (
	"io"
)

type Transcriber interface {
	Transcribe(r io.Reader) (string, error)
	TranscribeFile(name string) (string, error)
}
