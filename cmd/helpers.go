package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// getDBPath returns the path to the SQLite database, creating the parent
// directory if it does not exist.
func getDBPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".", "data.db")
	}
	dir := filepath.Join(cacheDir, "oh-my-learner")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return filepath.Join(".", "data.db")
	}
	return filepath.Join(dir, "data.db")
}

// findSubjectPack locates a subject pack TOML file for the given subject ID.
// It searches in order:
//  1. subjects/{id}.toml relative to CWD,
//  2. subjects/{id}.toml relative to the binary's directory.
func findSubjectPack(subjectID string) (string, error) {
	subjectID = strings.ToLower(subjectID)
	paths := []string{
		filepath.Join("subjects", subjectID+".toml"),
		filepath.Join(filepath.Dir(os.Args[0]), "subjects", subjectID+".toml"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("subject pack '%s' not found", subjectID)
}

// stdinReader is a package-level buffered reader so readLine calls share the
// same buffer. Without this, each readLine creates a new bufio.Reader that
// reads up to 4096 bytes from stdin on its first call, draining the pipe
// for all subsequent calls and causing EOF errors on piped input.
var stdinReader *bufio.Reader

func initStdinReader() {
	if stdinReader == nil {
		stdinReader = bufio.NewReader(os.Stdin)
	}
}

// readLine reads a single line from stdin, stripping the trailing newline.
// Must NOT create a new bufio.Reader per call — that drains the pipe.
func readLine() (string, error) {
	initStdinReader()
	s, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}

// getDailyReviewLimit reads the daily review limit from config.
// Defaults to 50 if config is missing or unparseable.
func getDailyReviewLimit() int {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 50
	}
	configPath := filepath.Join(homeDir, ".config", "oh-my-learner", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 50
	}
	var cfg struct {
		DailyReviewLimit int `toml:"daily_review_limit"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return 50
	}
	if cfg.DailyReviewLimit <= 0 {
		return 50
	}
	return cfg.DailyReviewLimit
}

// getConfigDir returns the oh-my-learner config directory path.
func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "oh-my-learner")
	}
	return filepath.Join(homeDir, ".config", "oh-my-learner")
}
