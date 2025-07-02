# 🎵 Cordelia - Chord Identifier Utility Specification v0.4

## 0. Document History

| Version | Date       | Author | Changes                                                                                                                                                             |
|---------|------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| v0.1    | (initial)  | User   | Initial Proof of Concept (POC) draft.                                                                                                                               |
| v0.2    | 2024-07-29 | Gemini | Refined logic for matching, error handling, and output.                                                                                                             |
| v0.3    | 2024-07-29 | Gemini | Integrated finer points on sorting, casing, and handling of empty/whitespace lines and empty files. Refined error handling table for clarity.                      |
| v0.4    | 2025-07-01 | Gemini | Renamed project to "Cordelia". Enhanced `--keys` flag to work on CLI arguments (chord names) without requiring `--batch`. Added chord name parsing logic to spec. |

## 1. Overview

**Cordelia** is a command-line Go program to identify possible chord names from a set of input notes and optionally estimate likely musical keys from a collection of chords. This specification details the behavior, features, and error handling for the utility.

---

## 2. Features

* **Note-to-Chord Identification**: Identifies chords from a list of individual notes.
* **Chord-to-Key Estimation**: Estimates the most likely musical key from a series of chord names or from notes in a batch file.
* **Inversion Detection**: An optional flag (`--inversions`) allows the tool to treat each note in a set as a potential root.
* **Batch Processing**: A `--batch` flag processes multiple chords (one per line, notes-based) from a file.

---

## 3. Command-Line Interface (CLI)

### **Usage**

```bash
# Identify a single chord from notes
cordelia [flags] <note1> <note2> ...

# Estimate key from a series of chord names
cordelia --keys <chord1> <chord2> ...

# Batch processing from a file
cordelia --batch <file> [flags]
```

### **Flags**

| Flag           | Argument Type | Description                                                                                                                                                           |
|----------------|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--notes`      | `string`      | Comma-separated list of notes (e.g., `"C,E,G,Bb"`). For identifying a single chord.                                                                                   |
| `--inversions` | `bool`        | When identifying a chord from notes, enables inversion detection by treating each note as a potential root.                                                           |
| `--batch`      | `string`      | Path to a file containing multiple chords (one chord per line, notes-based).                                                                                          |
| `--keys`       | `bool`        | Enables key estimation. When used with `--batch`, analyzes all notes in the file. When used without `--batch`, it analyzes chord names provided as positional arguments. |
| `--verbose`    | `bool`        | If present, shows detailed matching logic, including failed checks against the chord dictionary.                                                                      |
| `--help`       | `bool`        | If present, displays usage information and exits.                                                                                                                     |

---

## 4. Input & Parsing Logic

### **Note Parsing (Chord Identification Mode)**

* When not in key-estimation-only mode, arguments are treated as individual notes.
* **Supported Note Formats**: `A B C D E F G` with accidentals `#` (sharp) and `b` (flat). Case-insensitive.
* **Enharmonic Equivalence**: Notes are normalized internally (e.g., C# and Db are the same), but output preserves the original spelling of the root note.
* **Duplicate Notes**: Duplicates in an input set are ignored.

### **Chord Name Parsing (Key Estimation from Arguments)**

* This mode is active when `--keys` is used **without** the `--batch` flag.
* Positional arguments are treated as **chord names** (e.g., `C`, `Am`, `G7`, `F#m7`).
* The parser must identify:
    1.  The **root note** (e.g., `C`, `G`, `F#`).
    2.  The **chord quality/formula** (e.g., Major (default), `m`, `7`, `maj7`, `m7`).
* The application will generate the constituent notes for each parsed chord name based on the interval formulas in the chord dictionary.
    * *Example: "Am7"* -> Root: A, Quality: Minor 7th -> Intervals `[0, 3, 7, 10]` relative to A -> Generates notes: A, C, E, G.
* All generated notes from all chord arguments are aggregated for key estimation.

---

## 5. Core Logic

### **Note-to-Chord Identification**

* **Algorithm**: Uses **subset matching**. An input set of notes matches a chord if it contains all the intervals required by that chord's formula.
* **Dictionary**: Uses an internal dictionary mapping chord names to interval formulas (e.g., Major Triad: `[0, 4, 7]`).

### **Key Estimation**

* **Aggregation**: All notes (either from a batch file or generated from chord name arguments) are collected.
* **Comparison**: The aggregated notes are compared against all 12 Major and 12 Natural Minor scales.
* **Ranking**: Keys are ranked by the number of matching notes. Ties are broken alphabetically by key name.

---

## 6. Output Formats

### **Standard Chord Identification**

```
Input Notes: C E G
Root: C
Intervals: [0, 4, 7]
Matched Chords:
 - C Major Triad
```

### **Key Estimation from Arguments**

*Command:* `cordelia --keys C G Am F`
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

---

## 7. Error Handling & Exit Codes

* Error handling rules from v0.3 remain.
* New errors will be added for invalid chord name parsing (e.g., `cordelia --keys Bm#9`).

