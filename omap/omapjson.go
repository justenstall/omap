package omap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

var (
	_ json.Marshaler   = (*Map[string, any])(nil)
	_ json.Unmarshaler = (*Map[string, any])(nil)
)

// MarshalJSON implements [json.Marshaler].
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte(`null`), nil
	}
	buf := new(bytes.Buffer)
	// Opening bracket
	buf.WriteByte('{')
	first := true
	for key, value := range m.All() {
		// Marshal the key
		keyJSON, err := marshalKey(key)
		if err != nil {
			return nil, fmt.Errorf("marshalling key (type %T): %w", key, err)
		}

		// Marshal the value
		valueJSON, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshalling value (type %T): %w", value, err)
		}

		// Write leading comma after first value
		if !first {
			buf.WriteByte(',')
		}

		// Write key and value joined by colon
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(valueJSON)

		// Mark as not the first
		first = false
	}
	// Closing bracket
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// UnmarshalJSON implements [json.Unmarshaler].
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	// Empty input
	if len(data) == 0 || bytes.Equal(data, []byte(`null`)) {
		return nil
	}
	// Input too short or not an object
	if len(data) < 2 || data[0] != '{' || data[len(data)-1] != '}' {
		return fmt.Errorf("cannot parse %s as JSON object", string(data))
	}

	// Create a JSON decoder to read within the object
	// data = data[1 : len(data)-1] // remove leading and trailing bytes
	r := bytes.NewReader(data)
	d := json.NewDecoder(r)

	// Consume the '{' delimiter
	if _, err := d.Token(); err != nil {
		return err
	}

	// Decode entries until complete
	for d.More() {
		var (
			key   K
			value V
		)

		// Decode the key as a string
		var keyString string
		if err := d.Decode(&keyString); err != nil {
			return fmt.Errorf("unmarshalling key string: %w", err)
		}
		// Parse the key as its native type
		if err := parseKey(keyString, &key); err != nil {
			return fmt.Errorf("parsing key as type %T: %w", key, err)
		}
		// Decode the value
		if err := d.Decode(&value); err != nil {
			return fmt.Errorf("unmarshalling value (type %T): %w", value, err)
		}

		// Set the value in the map
		m.Set(key, value)
	}

	return nil
}

// marshalKey marshals a key as a JSON string.
func marshalKey(key any) ([]byte, error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}
	if len(keyJSON) == 0 {
		// Return empty JSON string
		return []byte(`""`), nil
	}
	// Check first character of JSON representation
	switch keyJSON[0] {
	case '"':
		// JSON value is a string, return as-is
		return keyJSON, nil
	case '[', '{':
		// JSON value is an array or object, return error
		return nil, fmt.Errorf("unsupported key type: %T", key)
	default:
		// JSON value is a number, boolean, or null
		// Format the key as a string (JSON only supports string keys in mappings)
		keyJSON, err = json.Marshal(string(keyJSON))
		if err != nil {
			return nil, fmt.Errorf("formatting as string: %w", err)
		}
		return keyJSON, nil
	}
}

// parseKey parses a string into key.
// key must be a pointer to a type whose underyling type
// satisfies cmp.Ordered.
func parseKey(keyString string, key any) error {
	// Handle all types in cmp.Ordered
	switch typedKey := any(key).(type) {
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
		*float32, *float64:
		// Unmarshal as JSON for non-string types
		return json.Unmarshal([]byte(keyString), typedKey)
	case *string:
		// Store string and return
		*typedKey = keyString
		return nil
	default:
		// Fall back to using reflection to check the
		// underlying type of the key.
		v := reflect.ValueOf(key)
		if v.Kind() != reflect.Pointer {
			// Key must be a pointer
			return fmt.Errorf("unsupported key type: %T", key)
		}

		// We know that key is a pointer, so call Elem()
		// to get the element type.
		switch v.Elem().Kind() {
		// Numeric types
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			// Unmarshal as JSON for non-string types
			return json.Unmarshal([]byte(keyString), key)
		// String types
		case reflect.String:
			v.Elem().SetString(keyString)
			return nil
		default:
			return fmt.Errorf("unsupported key type: %T", key)
		}
	}
}

// // cutByte slices s around the first instance of sep,
// // returning the text before and after sep.
// // The found result reports whether sep appears in s.
// // If sep does not appear in s, cut returns s, nil, false.
// //
// // cutByte returns slices of the original slice s, not copies.
// func cutByte(s []byte, sep byte) (before, after []byte, found bool) {
// 	if i := bytes.IndexByte(s, sep); i >= 0 {
// 		return s[:i], s[i+1:], true
// 	}
// 	return s, nil, false
// }
