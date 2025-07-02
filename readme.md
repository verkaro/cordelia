# 🎵 Cordelia - A Command-Line Chord Identifier

`cordelia` is a command-line tool written in Go that identifies possible musical chords from a set of notes. It can also analyze a series of chords to estimate the most likely musical key. It's designed to be a fast, simple, and reliable utility for musicians and developers.

---

## ✨ Features

* **Chord Identification**: Identifies standard chords (major, minor, dominant 7th, etc.) from a given set of notes.
* **Key Estimation**: Estimates the most likely key from a sequence of chord names or from notes in a batch file.
* **Subset Matching**: Correctly identifies chords even when extra, non-chord tones are present (e.g., identifies "C Major Triad" from the notes `C E G D`).
* **Inversion Detection**: Use the `--inversions` flag to test every note as a potential root of the chord.
* **Batch Processing**: Analyze a file containing multiple chords (one per line) using the `--batch` flag.
* **Flexible Input**: Provide notes or chord names directly on the command line.
* **Verbose Mode**: Use `--verbose` to see a detailed breakdown of how the tool matched (or failed to match) against its internal chord dictionary.

---

## 🚀 Usage

### Prerequisites

* Go (version 1.18 or later)

### Running from Source

To run the program directly without compiling, use `go run`:

```bash
go run main.go -- [flags] [arguments...]
```

*(Note: The `--` is important to separate Go's flags from the application's flags).*

### Examples

**1. Identify a single chord from notes:**

```bash
go run main.go -- C E G
```

*Output:*
```
Input Notes: C E G
Root: C
Intervals: [0, 4, 7]
Matched Chords:
 - C Major Triad
```

**2. Estimate the key from a chord progression:**

```bash
go run main.go -- --keys C G Am F
```

*Output:*
```
Processing Chords: C G Am F
Aggregated Notes: C D E F G A

Likely Keys:
 C Major (6 matches)
 A Minor (5 matches)
 F Major (5 matches)
 G Major (5 matches)
 ...
```

**3. Identify an inverted chord from notes:**

```bash
go run main.go -- --inversions E G C
```

*Output (will show results for C as the matching root):*
```
...
Input Notes: E G C
Root: C
Intervals: [0, 4, 7]
Matched Chords:
 - C Major Triad
```

**4. Process a batch file of notes and estimate the key:**

Create a file named `chords.txt`:
```
C G E
D A F#
G D B
```

Run the command:
```bash
go run main.go -- --batch chords.txt --keys
```

*Output:*
```
Processing chords.txt...
[1] C G E -> C Major Triad
[2] D A F# -> D Major Triad
[3] G D B -> G Major Triad
---
Key Estimation Results
Aggregated Notes: C D E F# G A B

Likely Keys:
 G Major (7 matches)
 A Minor (6 matches)
 C Major (6 matches)
 ...
```

---

## ⚙️ Command-Line Flags

| Flag           | Description                                                                                                                                                           |
|----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--notes`      | Comma-separated list of notes (e.g., `"C,E,G,Bb"`). For identifying a single chord.                                                                                   |
| `--inversions` | When identifying a chord from notes, enables inversion detection by treating each note as a potential root.                                                           |
| `--batch`      | Path to a file containing multiple chords (one per line, notes-based).                                                                                                |
| `--keys`       | Enables key estimation. When used with `--batch`, analyzes all notes in the file. When used without `--batch`, it analyzes chord names provided as positional arguments. |
| `--verbose`    | Show detailed matching logic, including failed checks against the dictionary.                                                                                         |
| `--help`       | Display usage information.                                                                                                                                            |

---

## 🛠️ Building and Testing

### Build from Source

To create a standalone executable:

```bash
go build -o cordelia main.go
```

You can then run the tool directly:
```bash
./cordelia C E G
```

### Running Tests

The project includes a comprehensive test suite that validates the core logic and the command-line interface.

```bash
go test -v
```

To run with the race detector:
```bash
go test -race -v
```

---

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## 🙏 Credits

* **Project Owner / Lead Developer**: [Your Name/Handle Here]
* **Implementation Assistance**: Gemini, a large language model from Google.

