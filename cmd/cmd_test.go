package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestHelpIncludesCommands verifies that --help output lists all expected subcommands.
func TestHelpIncludesCommands(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"--help"})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("learn --help failed: %v", err)
	}

	output := buf.String()
	expected := []string{"review", "add", "status", "report", "hook"}
	for _, cmd := range expected {
		if !strings.Contains(output, cmd) {
			t.Errorf("help output missing command %q", cmd)
		}
	}
}

// TestRootCommandNoArgs verifies that running the root command with no arguments
// prints help text (the root command has no Run/RunE handler) and does not error.
func TestRootCommandNoArgs(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("learn (no args) failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "learn") {
		t.Errorf("expected help output containing 'learn', got: %s", output)
	}
}

// TestInvalidCommand verifies that an unknown subcommand returns an error.
func TestInvalidCommand(t *testing.T) {
	RootCmd.SetArgs([]string{"nonexistent-command-xyz"})

	err := RootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %v", err)
	}
}

// TestSubcommandHelp verifies that each core subcommand's --help flag parses
// without error and produces output mentioning the subcommand name.
func TestSubcommandHelp(t *testing.T) {
	subcommands := []string{"review", "add", "status", "report", "hook"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			RootCmd.SetOut(buf)
			RootCmd.SetArgs([]string{name, "--help"})

			err := RootCmd.Execute()
			if err != nil {
				t.Fatalf("%s --help failed: %v", name, err)
			}

			output := buf.String()
			if !strings.Contains(output, name) {
				t.Errorf("%s --help output missing command name %q", name, name)
			}
		})
	}
}
