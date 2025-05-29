package speaker

import (
	"io"
	"os"
)

type SpeechStream interface {
	io.ReadCloser
}

type TempFileSpeechStream struct {
	io.ReadCloser
	path string
}

func (ts *TempFileSpeechStream) Close() error {
	ts.ReadCloser.Close()
	os.Remove(ts.path)
	return nil
}

func NewTempFileSpeechStream(path string) (*TempFileSpeechStream, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &TempFileSpeechStream{
		ReadCloser: f,
		path:       path,
	}, nil
}
