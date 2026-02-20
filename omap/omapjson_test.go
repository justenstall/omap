package omap

import (
	"cmp"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T {
	return &v
}

func testUnmarshal[K cmp.Ordered, V any](t *testing.T, data string, want *Map[K, V], wantErrString string) {
	t.Helper()
	got := &Map[K, V]{}
	err := json.Unmarshal([]byte(data), &got)
	if wantErrString != "" {
		assert.ErrorContains(t, err, wantErrString, "json.Unmarshal() error")
	} else {
		assert.NoError(t, err, "json.Unmarshal() error")
	}
	assert.Equal(t, want, got, "json.Unmarshal() output")
}

func TestUnmarshal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := `{"3":"value3","2":"value2","1":"value1"}`
		want := New([]Entry[string, string]{
			{"3", "value3"},
			{"2", "value2"},
			{"1", "value1"},
		}...)
		testUnmarshal(t, data, want, "")
	})
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
			wantKey:   nil,
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
			if tt.wantErr {
				assert.Error(t, err, "parseKey() error")
			} else {
				assert.NoError(t, err, "parseKey() error")
			}
			assert.Equal(t, tt.wantKey, tt.key, "parseKey() output")
		})
	}
}

type intWrapper int
type stringWrapper string
