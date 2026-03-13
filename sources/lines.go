package sources

import (
	"bufio"
	"io"
	"iter"
	"os"
)

// Lines returns a lazy sequence of lines from any io.Reader.
// Reads one line at a time.
//
// example:
//
//	file, _ := os.Open("app.log")
//	defer file.Close()
//
//	for line := range sources.Lines(file) {
//	    fmt.Println(line)
//	}
func Lines(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}
	}
}

// FileLines opens a file and returns a lazy sequence of its lines.
// Closes the file automatically when the sequence is exhausted
// or the caller stops early.
//
// example:
//
//	for line := range sources.FileLines("app.log") {
//	    fmt.Println(line)
//	}
func FileLines(path string) iter.Seq[string] {
	return func(yield func(string) bool) {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}
	}
}

// Stdin returns a lazy sequence of lines from standard input.
// Useful for piping data into a CLI tool.
//
// example:
//
//	for line := range sources.Stdin() {
//	    fmt.Println(line)
//	}
func Stdin() iter.Seq[string] {
	return Lines(os.Stdin)
}