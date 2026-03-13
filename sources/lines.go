package sources

import (
	"bufio"
	"context"
	"io"
	"iter"
	"os"
)

// Lines returns a lazy sequence of lines from any io.Reader.
func Lines(ctx context.Context, r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for {
			if ctx.Err() != nil {
				return
			}
			if !scanner.Scan() {
				break
			}
			if !yield(scanner.Text()) {
				return
			}
		}

		if scanner.Err() != nil {
			return
		}
	}
}

// LinesWithError is like Lines but surfaces read errors and oversized lines.
func LinesWithError(ctx context.Context, r io.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for {
			if ctx.Err() != nil {
				return
			}
			if !scanner.Scan() {
				break
			}
			if !yield(scanner.Text(), nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if ctx.Err() != nil {
				return
			}
			yield("", err)
		}
	}
}

// FileLines opens a file and returns a lazy sequence of its lines.
func FileLines(ctx context.Context, path string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if ctx.Err() != nil {
			return
		}

		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for {
			if ctx.Err() != nil {
				return
			}
			if !scanner.Scan() {
				break
			}
			if !yield(scanner.Text()) {
				return
			}
		}

		if scanner.Err() != nil {
			return
		}
	}
}

// FileLinesWithError is like FileLines but surfaces errors.
func FileLinesWithError(ctx context.Context, path string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if ctx.Err() != nil {
			return
		}

		file, err := os.Open(path)
		if err != nil {
			yield("", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for {
			if ctx.Err() != nil {
				return
			}
			if !scanner.Scan() {
				break
			}
			if !yield(scanner.Text(), nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if ctx.Err() != nil {
				return
			}
			yield("", err)
		}
	}
}

// Stdin returns a lazy sequence of lines from standard input.
func Stdin(ctx context.Context) iter.Seq[string] {
	return Lines(ctx, os.Stdin)
}
