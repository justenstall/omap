package omap

import (
	"reflect"
	"testing"
)

func ptrTo[T any](v T) *T {
	return &v
}

func Test_parseKey(t *testing.T) {
	tests := []struct {
		name      string
		keyString string
		key       any
		wantKey   any
		wantErr   bool
	}{
		{
			name:      "int",
			keyString: "100",
			key:       new(int),
			wantKey:   ptrTo(int(100)),
			wantErr:   false,
		},
		{
			name:      "int8",
			keyString: "100",
			key:       new(int8),
			wantKey:   ptrTo(int8(100)),
			wantErr:   false,
		},
		{
			name:      "int8 overflow",
			keyString: "100000000",
			key:       new(int8),
			wantKey:   ptrTo(int8(0)),
			wantErr:   true,
		},
		{
			name:      "float64",
			keyString: "100.05",
			key:       new(float64),
			wantKey:   ptrTo(float64(100.05)),
			wantErr:   false,
		},
		{
			name:      "string",
			keyString: "[1,2,3]",
			key:       new(string),
			wantKey:   ptrTo("[1,2,3]"),
			wantErr:   false,
		},
		{
			name:      "string double pointer",
			keyString: "[1,2,3]",
			key:       new(*string),
			wantKey:   new(*string),
			wantErr:   true,
		},
		{
			name:      "non pointer",
			keyString: "100",
			key:       int(0),
			wantKey:   int(0),
			wantErr:   true,
		},
		{
			name:      "int wrapper",
			keyString: "100",
			key:       new(intWrapper),
			wantKey:   ptrTo(intWrapper(100)),
			wantErr:   false,
		},
		{
			name:      "string wrapper",
			keyString: "100",
			key:       new(stringWrapper),
			wantKey:   ptrTo(stringWrapper("100")),
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseKey(tt.keyString, tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseKey() error: wantErr=%t, err=%+v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(tt.wantKey, tt.key) {
				t.Fatalf("parseKey() output: want=%+v, got=%+v", tt.wantKey, tt.key)
			}
		})
	}
}

type intWrapper int
type stringWrapper string
