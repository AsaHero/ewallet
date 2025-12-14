package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertOggToMp3(ctx context.Context, oggPath string) (string, error) {
	outPath := strings.TrimSuffix(oggPath, filepath.Ext(oggPath)) + ".mp3"

	// -ac 1 (mono) is perfect for voice
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-i", oggPath,
		"-ac", "1",
		"-b:a", "128k",
		outPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v, output=%s", err, string(out))
	}

	// optional: ensure file exists
	if _, statErr := os.Stat(outPath); statErr != nil {
		return "", fmt.Errorf("mp3 not created: %w", statErr)
	}

	return outPath, nil
}
