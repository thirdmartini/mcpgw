package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

func GetOutboundIP() string {
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			break
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return "localhost"
}

func WriteFile(filename string, r io.Reader) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func (s *Server) Transcribe(r io.Reader) (string, error) {
	if s.transcriber == nil {
		return "", errors.New("no transcriber")
	}

	tmid := uuid.New().String()

	tmpFile := path.Join(os.TempDir(), fmt.Sprintf("rasputin.%s", tmid))
	webmFile := tmpFile + ".mp4"
	wavFile := tmpFile + ".wav"
	defer os.Remove(webmFile)
	defer os.Remove(wavFile)
	log.Debugf("Transcribing %s -> %s\n", webmFile, wavFile)
	err := WriteFile(webmFile, r)
	if err != nil {
		return "", err
	}

	out, err := exec.Command(
		"./tools/ffmpeg",
		"-y", //overwrite
		"-i", webmFile,
		"-c:a", "pcm_s16le",
		"-ar", "16000",
		"-ac", "2",
		wavFile,
	).CombinedOutput()
	if err != nil {
		log.Debugf("FFMPEG:%s\n", string(out))
		return "", err
	}

	// exec.Command("/usr/bin/afplay", wavFile).Run()

	return s.transcriber.TranscribeFile(wavFile)
}

func MatchPrefix(s string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return prefix
		}
	}
	return ""
}
