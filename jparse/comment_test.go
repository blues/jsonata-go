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
			name:  "basic multiline comment",
			input: "/* this is a comment */ 42",
			want:  []tokenType{typeNumber},
		},
		{
			name:  "multiline comment between tokens",
			input: "1 /* comment */ + 2",
			want:  []tokenType{typeNumber, typePlus, typeNumber},
		},
		{
			name:  "inline comment",
			input: "42 // this is a comment\n43",
			want:  []tokenType{typeNumber, typeNumber},
		},
		{
			name:  "inline comment at end",
			input: "42 // this is a comment",
			want:  []tokenType{typeNumber},
		},
		{
			name:  "multiple inline comments",
			input: "1 // first\n2 // second\n3",
			want:  []tokenType{typeNumber, typeNumber, typeNumber},
		},
		{
			name:  "mixed comment styles",
			input: "1 /* multi */ 2 // inline\n3",
			want:  []tokenType{typeNumber, typeNumber, typeNumber},
		},
		{
			name:  "comment in object",
			input: "{/* comment */\"a\":1}",
			want:  []tokenType{typeBraceOpen, typeString, typeColon, typeNumber, typeBraceClose},
		},
		{
			name:  "comment in array",
			input: "[1,/* comment */2]",
			want:  []tokenType{typeBracketOpen, typeNumber, typeComma, typeNumber, typeBracketClose},
		},
		{
			name:    "unterminated multiline comment",
			input:   "/* unterminated",
			want:    []tokenType{typeError},
			wantErr: true,
		},
		{
			name:  "nested-looking comment",
			input: "/* outer /* inner */ 42",
			want:  []tokenType{typeNumber},
		},
		{
			name:  "complex nested comments",
			input: "/* a /* b /* c */ d */ e */ 42",
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
