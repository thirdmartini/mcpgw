package speaker

import (
	"os"
	"os/exec"
)

func WriteTempFile(data []byte) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), "tmpfile")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write(data)

	return f.Name(), err
}

func PlayAudioFile(path string) error {
	return exec.Command("/usr/bin/afplay", path).Run()
}

func PlayAudio(buffer []byte) error {
	fileName, err := WriteTempFile(buffer)
	if err != nil {
		return err
	}
	defer os.Remove(fileName)
	return PlayAudioFile(fileName)
}
