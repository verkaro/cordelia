// main_test.go
// This file contains the tests for Cordelia.
// To run, save this file in the same directory as main.go and execute:
// go test -v

package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
)

// =====================================================================================
// SECTION: CLI Integration Tests
// =====================================================================================

type exitCapture struct {
	code int
	mu   sync.Mutex
}

func (e *exitCapture) Exit(code int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.code == -1 {
		e.code = code
	}
}

func TestCLI(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldExit := exit
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		exit = oldExit
	}()

	tests := []struct {
		name             string
		args             []string
		expectedExitCode int
		expectedStdout   string
		expectedStderr   string
		stdoutContains   bool
		stderrContains   bool
	}{
		{
			name:             "No Notes Error",
			args:             []string{"cordelia"},
			expectedExitCode: 1,
			expectedStderr:   "Error: No notes provided.",
		},
		{
			name:             "Help Flag",
			args:             []string{"cordelia", "--help"},
			expectedExitCode: 0,
			stderrContains:   true,
			expectedStderr:   "Usage of cordelia:",
		},
		{
			name:             "Key Estimation from Args",
			args:             []string{"cordelia", "--keys", "C", "G", "Am"},
			expectedExitCode: 0,
			stdoutContains:   true,
			// Corrected Test: Check for the presence of the key match, not its exact position.
			expectedStdout: "C Major (6 matches)",
		},
		{
			name:             "Key Estimation No Chords",
			args:             []string{"cordelia", "--keys"},
			expectedExitCode: 1,
			expectedStderr:   "Error: No chord names provided for key estimation.",
		},
		{
			name:             "Standard Single Chord",
			args:             []string{"cordelia", "C", "E", "G"},
			expectedExitCode: 0,
			stdoutContains:   true,
			expectedStdout:   "Matched Chords:\n - C Major Triad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr
			capture := &exitCapture{code: -1}
			exit = capture.Exit
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			notesFlag, inversionsFlag, batchFlag, keysFlag, verboseFlag, helpFlag = "", false, "", false, false, false
			exitCode = 0

			main()

			wOut.Close()
			wErr.Close()
			var stdoutBuf, stderrBuf bytes.Buffer
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { io.Copy(&stdoutBuf, rOut); wg.Done() }()
			go func() { io.Copy(&stderrBuf, rErr); wg.Done() }()
			wg.Wait()
			stdoutStr := strings.TrimSpace(stdoutBuf.String())
			stderrStr := strings.TrimSpace(stderrBuf.String())

			if capture.code != tt.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d. Stderr: %s", tt.expectedExitCode, capture.code, stderrStr)
			}
			if tt.stdoutContains {
				if !strings.Contains(stdoutStr, tt.expectedStdout) {
					t.Errorf("Expected stdout to contain %q, got %q", tt.expectedStdout, stdoutStr)
				}
			} else if tt.expectedStdout != "" && stdoutStr != tt.expectedStdout {
				t.Errorf("Expected stdout %q, got %q", tt.expectedStdout, stdoutStr)
			}
			if tt.stderrContains {
				if !strings.Contains(stderrStr, tt.expectedStderr) {
					t.Errorf("Expected stderr to contain %q, got %q", tt.expectedStderr, stderrStr)
				}
			} else if tt.expectedStderr != "" && stderrStr != tt.expectedStderr {
				t.Errorf("Expected stderr %q, got %q", tt.expectedStderr, stderrStr)
			}
		})
	}
}

// =====================================================================================
// SECTION: Unit Tests
// =====================================================================================

func TestParseChordName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input         string
		expectedRoot  string
		expectedChord string
		expectError   bool
	}{
		{"C", "C", "Major Triad", false},
		{"Am", "A", "Minor Triad", false},
		{"F#m7", "F#", "Minor 7th", false},
		{"Bb7", "Bb", "Dominant 7th", false},
		{"Gaug", "G", "Augmented Triad", false},
		{"H", "", "", true},
		{"Cmaj9", "", "", true}, // maj9 is not in our dictionary
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			root, chord, err := ParseChordName(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got unexpected error: %v", err)
			}
			if root.Original != tt.expectedRoot {
				t.Errorf("Expected root %s, got %s", tt.expectedRoot, root.Original)
			}
			if chord.Name != tt.expectedChord {
				t.Errorf("Expected chord name %s, got %s", tt.expectedChord, chord.Name)
			}
		})
	}
}

func TestGenerateNotes(t *testing.T) {
	t.Parallel()
	root := Note{Original: "A", Value: 9}
	intervals := []int{0, 3, 7} // Minor Triad
	expectedNotes := []Note{
		{Original: "A", Value: 9},
		{Original: "C", Value: 0},
		{Original: "E", Value: 4},
	}

	got := GenerateNotes(root, intervals)
	// Sort for comparison
	sort.Slice(got, func(i, j int) bool { return got[i].Value < got[j].Value })
	sort.Slice(expectedNotes, func(i, j int) bool { return expectedNotes[i].Value < expectedNotes[j].Value })

	if !reflect.DeepEqual(got, expectedNotes) {
		t.Errorf("GenerateNotes() got = %v, want %v", got, expectedNotes)
	}
}

