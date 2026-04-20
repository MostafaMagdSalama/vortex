package iterx_test

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"testing"

	iterx "github.com/MostafaMagdSalama/vortex/iterx"
)

func TestFlatMap(t *testing.T) {
	type Order struct{ name string }
	type User struct{ orders []Order }

	tests := []struct {
		name    string
		users   []User
		usersIter iter.Seq2[User , error]
		buildFn func(User) iter.Seq2[Order, error]
		want    []string
		wantErr bool
	}{
		{
			name:  "empty input",
			users: []User{},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return seqToSeq2(slices.Values(u.orders))
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "normal case",
			users: []User{
				{orders: []Order{{"iphone"}, {"macbook"}}},
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return seqToSeq2(slices.Values(u.orders))
			},
			want:    []string{"iphone", "macbook"},
			wantErr: false,
		},
		{
			name: "inner error",
			users: []User{
				{orders: []Order{{"iphone"}}},
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return func(yield func(Order, error) bool) {
					yield(Order{}, fmt.Errorf("inner error"))
				}
			},
			wantErr: true,
		},
		{
			name: "outer error",
			usersIter:func(yield func(User, error)bool)  {
				 yield(User{}, fmt.Errorf("outer Error"))
			},
			buildFn: func(u User) iter.Seq2[Order, error] {
				return func(yield func(Order, error) bool) {
					yield(Order{}, fmt.Errorf("inner error"))
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string

			usersIter := tt.usersIter
			if usersIter == nil {
				usersIter = seqToSeq2(slices.Values(tt.users))
			}

			orderIter := iterx.FlatMap(context.Background(), usersIter, tt.buildFn)

			err := iterx.Drain(context.Background(), orderIter, func(o Order) error {
				result = append(result, o.name)
				return nil
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(result, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}
