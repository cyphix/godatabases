package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// JSONB wraps json.RawMessage for Postgres jsonb columns.
// Use with GORM: `gorm:"type:jsonb"`.
type JSONB json.RawMessage

// JSONBOf is a typed jsonb column backed by JSONB storage.
type JSONBOf[T any] struct {
	V T
}

var (
	errNilJSONB   = errors.New("types: nil JSONB")
	errInvalidJSON = errors.New("types: invalid JSON")
)

// Bytes returns a copy of the raw JSON bytes.
func (j JSONB) Bytes() []byte {
	if j == nil {
		return nil
	}
	return []byte(j)
}

// MarshalJSON implements json.Marshaler.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errNilJSONB
	}
	if len(data) == 0 {
		*j = JSONB([]byte("null"))
		return nil
	}
	*j = JSONB(append(json.RawMessage(nil), data...))
	return nil
}

// Value implements driver.Valuer for Postgres jsonb.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	if !json.Valid(j) {
		return nil, fmt.Errorf("%w: value is not valid JSON", errInvalidJSON)
	}
	return []byte(j), nil
}

// Scan implements sql.Scanner for Postgres jsonb.
func (j *JSONB) Scan(src any) error {
	if j == nil {
		return errNilJSONB
	}
	switch v := src.(type) {
	case nil:
		*j = nil
		return nil
	case []byte:
		if len(v) == 0 {
			*j = JSONB([]byte("null"))
			return nil
		}
		if !json.Valid(v) {
			return fmt.Errorf("%w: scanned bytes are not valid JSON", errInvalidJSON)
		}
		*j = JSONB(append(json.RawMessage(nil), v...))
		return nil
	case string:
		return j.Scan([]byte(v))
	default:
		return fmt.Errorf("types: cannot scan %T into JSONB", src)
	}
}

// NewJSONB marshals v into JSONB.
func NewJSONB(v any) (JSONB, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSONB(data), nil
}

// Unmarshal decodes JSONB into dst.
func (j JSONB) Unmarshal(dst any) error {
	if j == nil {
		return json.Unmarshal([]byte("null"), dst)
	}
	return json.Unmarshal(j, dst)
}

// Value implements driver.Valuer for typed jsonb columns.
func (j JSONBOf[T]) Value() (driver.Value, error) {
	raw, err := NewJSONB(j.V)
	if err != nil {
		return nil, err
	}
	return raw.Value()
}

// Scan implements sql.Scanner for typed jsonb columns.
func (j *JSONBOf[T]) Scan(src any) error {
	var raw JSONB
	if err := raw.Scan(src); err != nil {
		return err
	}
	if raw == nil {
		var zero T
		j.V = zero
		return nil
	}
	return raw.Unmarshal(&j.V)
}

// MarshalJSON implements json.Marshaler.
func (j JSONBOf[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.V)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONBOf[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.V)
}
