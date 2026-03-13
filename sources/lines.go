package sources

import (
	"bufio"
	"io"
	"iter"
	"os"
)

// Lines returns a lazy sequence of lines from any io.Reader.
// Stops and surfaces errors via LinesWithError if you need error handling.
func Lines(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		scanner := bufio.NewScanner(r)

		// increase buffer to handle lines larger than 64KB
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max line

		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}

		// scanner.Err() is nil on normal EOF
		// non-nil means something actually went wrong
		if err := scanner.Err(); err != nil {
			// Lines() is silent — use LinesWithError if you need errors
			return
		}
	}
}

// LinesWithError is like Lines but surfaces read errors and oversized lines.
// Use this when processing untrusted input or large files.
//
// example:
//
//	for line, err := range sources.LinesWithError(file) {
//	    if err != nil {
//	        log.Println("read error:", err)
//	        return
//	    }
//	    fmt.Println(line)
//	}
func LinesWithError(r io.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max line

		for scanner.Scan() {
			if !yield(scanner.Text(), nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield("", err) // surface the error to caller
		}
	}
}

// FileLines opens a file and returns a lazy sequence of its lines.
func FileLines(path string) iter.Seq[string] {
	return func(yield func(string) bool) {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			return
		}
	}
}

// FileLinesWithError is like FileLines but surfaces errors.
//
// example:
//
//	for line, err := range sources.FileLinesWithError("app.log") {
//	    if err != nil {
//	        log.Println("error:", err)
//	        return
//	    }
//	    fmt.Println(line)
//	}
func FileLinesWithError(path string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		file, err := os.Open(path)
		if err != nil {
			yield("", err) // surface open error
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			if !yield(scanner.Text(), nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield("", err) // surface scan error
		}
	}
}

// Stdin returns a lazy sequence of lines from standard input.
func Stdin() iter.Seq[string] {
	return Lines(os.Stdin)
}