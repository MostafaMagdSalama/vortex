package iterx_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/MostafaMagdSalama/vortex/iterx"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		input    []rune
		n        int
		expected [][]rune
	}{
		{
			name:     "even chunks",
			input:    []rune{'a', 'b', 'c', 'd'},
			n:        2,
			expected: [][]rune{{'a', 'b'}, {'c', 'd'}},
		},
		{
			name:     "uneven chunks",
			input:    []rune{'a', 'b', 'c', 'd', 'e'},
			n:        2,
			expected: [][]rune{{'a', 'b'}, {'c', 'd'}, {'e'}},
		},
		{
			name:     "chunk size larger than input",
			input:    []rune{'a', 'b'},
			n:        5,
			expected: [][]rune{{'a', 'b'}},
		},
		{
			name:     "chunk size equals input length",
			input:    []rune{'a', 'b', 'c'},
			n:        3,
			expected: [][]rune{{'a', 'b', 'c'}},
		},
		{
			name:     "chunk size of one",
			input:    []rune{'a', 'b', 'c'},
			n:        1,
			expected: [][]rune{{'a'}, {'b'}, {'c'}},
		},
		{
			name:     "empty input",
			input:    []rune{},
			n:        3,
			expected: [][]rune{},
		},
		{
			name:     "zero chunk size",
			input:    []rune{'a', 'b', 'c'},
			n:        0,
			expected: [][]rune{},
		},
		{
			name:     "negative chunk size",
			input:    []rune{'a', 'b', 'c'},
			n:        -1,
			expected: [][]rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result [][]rune

			inputIter := seqToSeq2(slices.Values(tt.input))

			chunkIter := iterx.Chunk(context.Background(), inputIter, tt.n)

			iterx.Drain(context.Background(), chunkIter, func(chunk []rune) error {
				result = append(result, chunk)
				return nil
			})

			if len(result) != len(tt.expected) {
				fmt.Println(result)
				t.Fatalf("expected %d chunks, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if !slices.Equal(result[i], tt.expected[i]) {
					t.Fatalf("chunk %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestChunk_PropagatesError(t *testing.T) {
	sentinelErr := errors.New("stream error")

	errAt := 3 // inject error after 3 elements

	inputIter := func(yield func(rune, error) bool) {
		input := []rune{'a', 'b', 'c', 'd', 'e'}
		for i, v := range input {
			if i == errAt {
				yield(0, sentinelErr)
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}

	var gotErr error

	chunkIter := iterx.Chunk(context.Background(), inputIter, 2)

	iterx.Drain(context.Background(), chunkIter, func(chunk []rune) error {
		return nil
	})

	// Drain directly to capture the error
	chunkIter(func(chunk []rune, err error) bool {
		if err != nil {
			gotErr = err
			return false
		}
		return true
	})

	if !errors.Is(gotErr, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", gotErr)
	}
}

func TestChunk_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	input := []rune{'a', 'b', 'c', 'd', 'e', 'f'}
	inputIter := seqToSeq2(slices.Values(input))

	chunkIter := iterx.Chunk(ctx, inputIter, 2)

	var result [][]rune
	chunkIter(func(chunk []rune, err error) bool {
		if err != nil {
			return false
		}
		result = append(result, chunk)
		return true
	})

	if len(result) > 0 {
		t.Fatalf("expected no results with cancelled context, got %d chunks", len(result))
	}
}

func TestChunk_StopsEarly(t *testing.T) {
	input := []rune{'a', 'b', 'c', 'd', 'e', 'f'}
	inputIter := seqToSeq2(slices.Values(input))

	chunkIter := iterx.Chunk(context.Background(), inputIter, 2)

	var result [][]rune
	chunkIter(func(chunk []rune, err error) bool {
		if err != nil {
			return false
		}
		result = append(result, chunk)
		return len(result) < 2 // stop after 2 chunks
	})

	if len(result) != 2 {
		t.Fatalf("expected 2 chunks after early stop, got %d", len(result))
	}
}