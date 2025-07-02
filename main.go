// main.go
// This file contains the complete implementation for Cordelia, a command-line
// chord and key identification utility.
// Version: 0.4
// To run, save this file and execute:
// go run main.go -- [args]

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
//	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// --- Global Variables ---
var (
	// These variables are set via command-line flags.
	notesFlag      string
	inversionsFlag bool
	batchFlag      string
	keysFlag       bool
	verboseFlag    bool
	helpFlag       bool

	// exit is a hook for testing to intercept calls to os.Exit.
	exit = os.Exit
)

// exitCode holds the final exit code of the program. It's updated on errors.
var exitCode = 0

// --- Main Function ---
// Entry point of the application.
func main() {
	// Setup and parse command-line flags.
	setupFlags()

	// Handle --help flag immediately.
	if helpFlag {
		flag.Usage()
		exit(0)
		return
	}

	// Validate flag dependencies.
	if err := validateFlags(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exit(1)
		return
	}

	// Determine the source of notes (flags vs. positional args).
	args := flag.Args()

	// Decide program mode based on flags.
	if keysFlag {
		if batchFlag != "" {
			// Key estimation from a batch file of notes.
			runBatchMode(batchFlag)
		} else {
			// Key estimation from CLI args (chord names).
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: No chord names provided for key estimation.")
				exit(1)
				return
			}
			runKeyEstimationFromArgs(args)
		}
	} else {
		// Single chord identification from notes.
		noteStrings, err := getNoteStringsFromInput(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			exit(1)
			return
		}
		if len(noteStrings) == 0 {
			fmt.Fprintln(os.Stderr, "Error: No notes provided.")
			exit(1)
			return
		}
		runSingleChordMode(noteStrings)
	}

	exit(exitCode)
}

// --- CLI & Program Flow ---

// setupFlags defines and configures the command-line flags.
func setupFlags() {
	flag.StringVar(&notesFlag, "notes", "", "Comma-separated list of notes (e.g., \"C,E,G,Bb\").")
	flag.BoolVar(&inversionsFlag, "inversions", false, "Enable inversion detection by treating each note as a potential root.")
	flag.StringVar(&batchFlag, "batch", "", "Path to a file containing multiple chords (one chord per line, notes-based).")
	flag.BoolVar(&keysFlag, "keys", false, "Enables key estimation.")
	flag.BoolVar(&verboseFlag, "verbose", false, "Show detailed matching logic, including failed checks.")
	flag.BoolVar(&helpFlag, "help", false, "Display usage information.")

	// Custom usage message to match the spec.
	flag.Usage = func() {
		appName := "cordelia"
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", appName)
		fmt.Fprintf(os.Stderr, "  Identify a chord from notes: %s [flags] <note1> <note2> ...\n", appName)
		fmt.Fprintf(os.Stderr, "  Estimate key from chords:    %s --keys <chord1> <chord2> ...\n", appName)
		fmt.Fprintf(os.Stderr, "  Batch processing from file:  %s --batch <file> [flags]\n", appName)
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()
}

// validateFlags checks for invalid combinations of flags.
func validateFlags() error {
	if !keysFlag && batchFlag != "" {
		// Allow batch mode without keys for just chord identification.
		return nil
	}
	// No invalid combinations to check in v0.4
	return nil
}

// getNoteStringsFromInput determines which notes to use based on flags and args.
func getNoteStringsFromInput(posArgs []string) ([]string, error) {
	// --notes flag takes precedence.
	if notesFlag != "" {
		if len(posArgs) > 0 {
			fmt.Fprintln(os.Stderr, "Warning: Both positional arguments and --notes provided; using --notes.")
		}
		return strings.Split(notesFlag, ","), nil
	}

	return posArgs, nil
}

// runKeyEstimationFromArgs handles the new mode for key estimation from chord names.
func runKeyEstimationFromArgs(chordNames []string) {
	var allNotes []Note
	fmt.Printf("Processing Chords: %s\n", strings.Join(chordNames, " "))

	for _, name := range chordNames {
		root, chordDef, err := ParseChordName(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not parse chord name '%s': %v\n", name, err)
			exitCode = 1
			return
		}

		chordNotes := GenerateNotes(root, chordDef.Intervals)
		allNotes = append(allNotes, chordNotes...)
	}

	printKeyEstimation(allNotes)
}

// runSingleChordMode processes a single set of notes for chord identification.
func runSingleChordMode(noteStrings []string) {
	notes, err := parseAndValidateNotes(noteStrings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitCode = 1
		return
	}

	rootsToTest := []Note{notes[0]}
	if inversionsFlag {
		rootsToTest = notes
	}

	for _, root := range rootsToTest {
		intervals := CalculateIntervals(root, notes)
		matches := FindMatches(intervals)

		if verboseFlag {
			printVerboseOutput(root, notes, intervals, matches)
		} else {
			printStandardOutput(root, notes, intervals, matches)
		}
	}
}

// runBatchMode processes a file line by line.
func runBatchMode(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", filename)
		exitCode = 1
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get file stats for %s\n", filename)
		exitCode = 1
		return
	}
	if stat.Size() == 0 {
		exit(0)
		return
	}

	fmt.Printf("Processing %s...\n", filename)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var allNotes []Note
	batchHasErrors := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			fmt.Fprintf(os.Stderr, "Error on line %d: No notes provided\n", lineNum)
			batchHasErrors = true
			continue
		}

		noteStrings := strings.Fields(line)
		notes, err := parseAndValidateNotes(noteStrings)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error on line %d: %v\n", lineNum, err)
			batchHasErrors = true
			continue
		}

		if keysFlag {
			allNotes = append(allNotes, notes...)
		}

		root := notes[0]
		intervals := CalculateIntervals(root, notes)
		matches := FindMatches(intervals)

		var matchStrings []string
		for _, m := range matches {
			isSubset := len(notes) > len(m.Intervals)
			matchStr := fmt.Sprintf("%s %s", root.Original, m.Name)
			if isSubset {
				matchStr += " (subset)"
			}
			matchStrings = append(matchStrings, matchStr)
		}

		if len(matchStrings) == 0 {
			fmt.Printf("[%d] %s -> No match found\n", lineNum, line)
		} else {
			fmt.Printf("[%d] %s -> %s\n", lineNum, line, strings.Join(matchStrings, ", "))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		exitCode = 1
		return
	}

	if keysFlag {
		printKeyEstimation(allNotes)
	}

	if batchHasErrors {
		exitCode = 2
	}
}

// --- Output Formatting ---

func printStandardOutput(root Note, notes []Note, intervals []int, matches []Match) {
	fmt.Printf("Input Notes: %s\n", SliceToString(notes))
	fmt.Printf("Root: %s\n", root.Original)
	fmt.Printf("Intervals: %v\n", intervals)
	fmt.Println("Matched Chords:")
	if len(matches) == 0 {
		fmt.Println(" - None")
	} else {
		for _, m := range matches {
			isSubset := len(notes) > len(m.Intervals)
			matchStr := fmt.Sprintf(" - %s %s", root.Original, m.Name)
			if isSubset {
				matchStr += " (subset)"
			}
			fmt.Println(matchStr)
		}
	}
	fmt.Println()
}

func printVerboseOutput(root Note, notes []Note, intervals []int, matches []Match) {
	fmt.Printf("Input Notes: %s\n", SliceToString(notes))
	fmt.Printf("Root: %s\n", root.Original)
	fmt.Printf("Input Intervals: %v\n", intervals)
	fmt.Println("---")
	fmt.Println("Checking Dictionary...")

	intervalSet := make(map[int]struct{})
	for _, i := range intervals {
		intervalSet[i] = struct{}{}
	}

	allChords := GetDictionary()
	for _, c := range allChords {
		match, reason := c.Check(intervals, intervalSet)
		if match {
			fmt.Printf("✅ Match: %s %v\n", c.Name, c.Intervals)
		} else {
			fmt.Printf("❌ No Match: %s %v (%s)\n", c.Name, c.Intervals, reason)
		}
	}

	fmt.Println("---")
	fmt.Println("Matched Chords:")
	if len(matches) == 0 {
		fmt.Println(" - None")
	} else {
		for _, m := range matches {
			isSubset := len(notes) > len(m.Intervals)
			matchStr := fmt.Sprintf(" - %s %s", root.Original, m.Name)
			if isSubset {
				matchStr += " (subset)"
			}
			fmt.Println(matchStr)
		}
	}
	fmt.Println()
}

func printKeyEstimation(allNotes []Note) {
	fmt.Println("---")
	fmt.Println("Key Estimation Results")

	uniqueNotes := Unique(allNotes)
	sort.Slice(uniqueNotes, func(i, j int) bool {
		return uniqueNotes[i].Value < uniqueNotes[j].Value
	})
	fmt.Printf("Aggregated Notes: %s\n\n", SliceToString(uniqueNotes))

	keyMatches := Estimate(uniqueNotes)
	if len(keyMatches) == 0 {
		fmt.Println("Could not determine likely keys.")
	} else {
		fmt.Println("Likely Keys:")
		for _, km := range keyMatches {
			fmt.Printf(" %s (%d matches)\n", km.Name, km.MatchCount)
		}
	}
}

// --- Utility Functions ---

func parseAndValidateNotes(noteStrings []string) ([]Note, error) {
	var notes []Note
	for _, s := range noteStrings {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		n, err := ParseNote(s)
		if err != nil {
			return nil, fmt.Errorf("invalid note '%s' in input", s)
		}
		notes = append(notes, n)
	}

	if len(notes) == 0 {
		return nil, errors.New("no valid notes provided")
	}

	return Unique(notes), nil
}

// =====================================================================================
// SECTION: Note Logic
// =====================================================================================

type Note struct {
	Original string
	Value    int
}

var noteMap = map[string]int{
	"C": 0, "B#": 0, "C#": 1, "DB": 1, "D": 2, "D#": 3, "EB": 3, "E": 4, "FB": 4,
	"F": 5, "E#": 5, "F#": 6, "GB": 6, "G": 7, "G#": 8, "AB": 8, "A": 9, "A#": 10, "BB": 10,
	"B": 11, "CB": 11,
}

var valueToName = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func ParseNote(s string) (Note, error) {
	if s == "" {
		return Note{}, fmt.Errorf("cannot parse empty string")
	}
	runes := []rune(s)
	normalized := string(unicode.ToTitle(runes[0]))
	if len(runes) > 1 {
		normalized += strings.ToUpper(string(runes[1:]))
	}
	normalized = strings.Replace(normalized, "b", "B", 1)

	value, ok := noteMap[normalized]
	if !ok {
		return Note{}, fmt.Errorf("unrecognized note")
	}
	return Note{Original: s, Value: value}, nil
}

func Unique(notes []Note) []Note {
	seen := make(map[int]struct{})
	var uniqueNotes []Note
	for _, n := range notes {
		if _, ok := seen[n.Value]; !ok {
			seen[n.Value] = struct{}{}
			uniqueNotes = append(uniqueNotes, n)
		}
	}
	return uniqueNotes
}

func SliceToString(notes []Note) string {
	var parts []string
	for _, n := range notes {
		if n.Original == "" {
			parts = append(parts, valueToName[n.Value])
		} else {
			parts = append(parts, n.Original)
		}
	}
	return strings.Join(parts, " ")
}

// =====================================================================================
// SECTION: Chord Logic
// =====================================================================================

type Chord struct {
	Name      string
	Suffixes  []string // Suffixes used for parsing chord names, e.g., "m", "min"
	Intervals []int
}

type Match struct {
	Name      string
	Intervals []int
}

var chordDictionary = []Chord{
	{Name: "Major 7th", Suffixes: []string{"maj7", "M7"}, Intervals: []int{0, 4, 7, 11}},
	{Name: "Minor-Major 7th", Suffixes: []string{"m(maj7)"}, Intervals: []int{0, 3, 7, 11}},
	{Name: "Minor 7th", Suffixes: []string{"m7", "min7"}, Intervals: []int{0, 3, 7, 10}},
	{Name: "Dominant 7th", Suffixes: []string{"7", "dom7"}, Intervals: []int{0, 4, 7, 10}},
	{Name: "Major Triad", Suffixes: []string{"", "M"}, Intervals: []int{0, 4, 7}},
	{Name: "Minor Triad", Suffixes: []string{"m", "min"}, Intervals: []int{0, 3, 7}},
	{Name: "Diminished Triad", Suffixes: []string{"dim"}, Intervals: []int{0, 3, 6}},
	{Name: "Augmented Triad", Suffixes: []string{"aug", "+"}, Intervals: []int{0, 4, 8}},
	{Name: "Sus2", Suffixes: []string{"sus2"}, Intervals: []int{0, 2, 7}},
	{Name: "Sus4", Suffixes: []string{"sus4"}, Intervals: []int{0, 5, 7}},
}

func GetDictionary() []Chord {
	return chordDictionary
}

// ParseChordName breaks a string like "F#m7" into a root note and a Chord definition.
func ParseChordName(name string) (Note, Chord, error) {
	// First, try to parse the longest possible note name (e.g., "C#", "Db").
	var rootNote Note
	var quality string
	var err error

	if len(name) > 1 {
		// Check for two-character note names like "C#" or "Db"
		if rootNote, err = ParseNote(name[:2]); err == nil {
			quality = name[2:]
		} else if rootNote, err = ParseNote(name[:1]); err == nil {
			quality = name[1:]
		} else {
			return Note{}, Chord{}, fmt.Errorf("invalid root note in chord name")
		}
	} else {
		if rootNote, err = ParseNote(name); err == nil {
			quality = ""
		} else {
			return Note{}, Chord{}, fmt.Errorf("invalid root note in chord name")
		}
	}

	// Now find the chord definition that matches the quality suffix.
	for _, chordDef := range chordDictionary {
		for _, suffix := range chordDef.Suffixes {
			if quality == suffix {
				return rootNote, chordDef, nil
			}
		}
	}

	return Note{}, Chord{}, fmt.Errorf("unknown chord quality: '%s'", quality)
}

func GenerateNotes(root Note, intervals []int) []Note {
	notes := make([]Note, len(intervals))
	for i, interval := range intervals {
		noteValue := (root.Value + interval) % 12
		// We don't have the "original" spelling, so we create a canonical one.
		notes[i] = Note{Original: valueToName[noteValue], Value: noteValue}
	}
	return notes
}

func CalculateIntervals(root Note, notes []Note) []int {
	var intervals []int
	for _, n := range notes {
		interval := n.Value - root.Value
		if interval < 0 {
			interval += 12
		}
		intervals = append(intervals, interval)
	}
	sort.Ints(intervals)
	return intervals
}

func (c Chord) Check(inputIntervals []int, inputSet map[int]struct{}) (bool, string) {
	if len(inputIntervals) < len(c.Intervals) {
		return false, fmt.Sprintf("requires %d intervals, input has %d", len(c.Intervals), len(inputIntervals))
	}
	for _, requiredInterval := range c.Intervals {
		if _, ok := inputSet[requiredInterval]; !ok {
			return false, fmt.Sprintf("missing interval %d", requiredInterval)
		}
	}
	return true, ""
}

func FindMatches(intervals []int) []Match {
	var matches []Match
	intervalSet := make(map[int]struct{})
	for _, i := range intervals {
		intervalSet[i] = struct{}{}
	}

	for _, chordDef := range chordDictionary {
		if ok, _ := chordDef.Check(intervals, intervalSet); ok {
			matches = append(matches, Match{Name: chordDef.Name, Intervals: chordDef.Intervals})
		}
	}
	return matches
}

// =====================================================================================
// SECTION: Key Logic
// =====================================================================================

type Key struct {
	Name  string
	Notes map[int]struct{}
}

type KeyMatch struct {
	Name       string
	MatchCount int
}

var keySignatures = []Key{}

func init() {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	flatNames := []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	majorPattern := []int{0, 2, 4, 5, 7, 9, 11}
	minorPattern := []int{0, 2, 3, 5, 7, 8, 10}

	for i := 0; i < 12; i++ {
		majorNotes := make(map[int]struct{})
		for _, interval := range majorPattern {
			majorNotes[(i+interval)%12] = struct{}{}
		}
		keySignatures = append(keySignatures, Key{Name: flatNames[i] + " Major", Notes: majorNotes})

		minorNotes := make(map[int]struct{})
		for _, interval := range minorPattern {
			minorNotes[(i+interval)%12] = struct{}{}
		}
		keySignatures = append(keySignatures, Key{Name: noteNames[i] + " Minor", Notes: minorNotes})
	}
}

func Estimate(notes []Note) []KeyMatch {
	if len(notes) == 0 {
		return nil
	}
	var matches []KeyMatch
	for _, keySig := range keySignatures {
		count := 0
		uniqueNotes := Unique(notes) // Ensure we only count each pitch class once
		for _, n := range uniqueNotes {
			if _, ok := keySig.Notes[n.Value]; ok {
				count++
			}
		}
		if count > 0 {
			matches = append(matches, KeyMatch{Name: keySig.Name, MatchCount: count})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].MatchCount != matches[j].MatchCount {
			return matches[i].MatchCount > matches[j].MatchCount
		}
		return matches[i].Name < matches[j].Name
	})
	return matches
}

