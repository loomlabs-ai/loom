package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	GroqBaseURL       string = "https://api.groq.com/openai/v1"
	OpenRouterBaseURL string = "https://openrouter.ai/api/v1"
)

var LoomPath string

func Init() error {
	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "config: failed to determine user home directory: %v\n", err)
		os.Exit(1)
	}

	LoomPath = filepath.Join(home, ".loom")

	return os.MkdirAll(LoomPath, 0755)
}

func GetLoomPath() string {
	return LoomPath
}
