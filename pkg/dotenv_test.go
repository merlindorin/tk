package pkg_test

import (
	"bytes"
	"testing"

	"github.com/merlindorin/tk/pkg"
)

func TestDotEnvMarshal(t *testing.T) {
	type args struct {
		s any
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "nominal",
			args: args{
				s: struct{ Foo string }{Foo: "hello world"},
			},
			wantW:   "Foo=\"hello world\"",
			wantErr: false,
		},
		{
			name: "ptr",
			args: args{
				s: &struct{ Foo string }{Foo: "hello world"},
			},
			wantW:   "Foo=\"hello world\"",
			wantErr: false,
		},
		{
			name: "multiple field",
			args: args{
				s: struct {
					Foo string
					Bar string
				}{
					Foo: "hello world",
					Bar: "goodbye world",
				},
			},
			wantW:   "Bar=\"goodbye world\"\nFoo=\"hello world\"",
			wantErr: false,
		},
		{
			name: "multiple field with ignored types",
			args: args{
				s: struct {
					Foo string
					Bar bool
				}{
					Foo: "hello world",
				},
			},
			wantW:   "Foo=\"hello world\"",
			wantErr: false,
		},
		{
			name: "field with tag",
			args: args{
				s: struct {
					Foo string `dot:"FOO"`
				}{Foo: "hello world"},
			},
			wantW:   "FOO=\"hello world\"",
			wantErr: false,
		},
		{
			name: "unsupported type",
			args: args{
				s: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name:    "unsupported nil value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := pkg.DotEnvMarshal(tt.args.s, w)
			if (err != nil) != tt.wantErr {
				t.Errorf("DotEnvMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("DotEnvMarshal() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
