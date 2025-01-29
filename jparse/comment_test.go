package jparse

import (
	"testing"
)

func TestCommentLexer(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []tokenType
		wantErr bool
	}{
		{
			name:  "basic comment",
			input: "/* this is a comment */ 42",
			want:  []tokenType{typeNumber},
		},
		{
			name:  "comment between tokens",
			input: "1 /* comment */ + 2",
			want:  []tokenType{typeNumber, typePlus, typeNumber},
		},
		{
			name:    "unterminated comment",
			input:   "/* unterminated",
			want:    []tokenType{typeError},
			wantErr: true,
		},
		{
			name:  "nested-looking comment",
			input: "/* outer /* inner */ 42",
			want:  []tokenType{typeNumber},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			l := newLexer(tt.input)
			var got []tokenType
			var hasError bool
			
			for {
				tok := l.next(true)
				if tok.Type == typeEOF {
					break
				}
				if tok.Type == typeError {
					hasError = true
				}
				got = append(got, tok.Type)
			}

			if hasError != tt.wantErr {
				t.Errorf("got error = %v, want %v", hasError, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("token[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
