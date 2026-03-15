package sources

import (
	"bufio"
	"context"
	"io"
	"iter"
	"os"

	"github.com/MostafaMagdSalama/vortex"
)

// Lines returns a lazy sequence of lines from any io.Reader.
func Lines(ctx context.Context, r io.Reader) iter.Seq2[string, error] {
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
			yield("", vortex.Wrap("sources.Lines", err))
		}
	}
}

// FileLines opens a file and returns a lazy sequence of its lines.
func FileLines(ctx context.Context, path string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if ctx.Err() != nil {
			yield("", vortex.Wrap("sources.FileLines", ctx.Err()))
			return
		}

		file, err := os.Open(path)
		if err != nil {
			yield("", vortex.Wrap("sources.FileLines", err))
			return
		}
		defer file.Close()

		for line, err := range Lines(ctx, file) {
			if !yield(line, err) {
				return
			}
		}
	}
}


// Stdin returns a lazy sequence of lines from standard input.
func Stdin(ctx context.Context) iter.Seq2[string, error] {
	return Lines(ctx, os.Stdin)
}
