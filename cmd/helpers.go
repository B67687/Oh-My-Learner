package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// readLine reads a single line from stdin, stripping the trailing newline.
func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	s, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}
