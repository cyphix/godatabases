package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONBValueAndScan(t *testing.T) {
	t.Parallel()

	original := JSONB(`{"name":"worldbuilder"}`)
	value, err := original.Value()
	require.NoError(t, err)

	var scanned JSONB
	require.NoError(t, scanned.Scan(value))
	assert.JSONEq(t, string(original), string(scanned))
}

func TestJSONBScanNil(t *testing.T) {
	t.Parallel()

	var j JSONB
	require.NoError(t, j.Scan(nil))
	assert.Nil(t, j)
}

func TestJSONBScanString(t *testing.T) {
	t.Parallel()

	var j JSONB
	require.NoError(t, j.Scan(`{"ok":true}`))
	assert.JSONEq(t, `{"ok":true}`, string(j))
}

func TestJSONBInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := JSONB(`{invalid`).Value()
	require.Error(t, err)
}

func TestNewJSONBAndUnmarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		Count int `json:"count"`
	}

	raw, err := NewJSONB(payload{Count: 3})
	require.NoError(t, err)

	var decoded payload
	require.NoError(t, raw.Unmarshal(&decoded))
	assert.Equal(t, 3, decoded.Count)
}

func TestJSONBOfRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		Label string `json:"label"`
	}

	col := JSONBOf[payload]{V: payload{Label: "universe"}}
	value, err := col.Value()
	require.NoError(t, err)

	var scanned JSONBOf[payload]
	require.NoError(t, scanned.Scan(value))
	assert.Equal(t, "universe", scanned.V.Label)
}

func TestJSONBOfJSONMarshaling(t *testing.T) {
	t.Parallel()

	col := JSONBOf[map[string]string]{V: map[string]string{"kind": "lore"}}
	data, err := json.Marshal(col)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind":"lore"}`, string(data))

	var decoded JSONBOf[map[string]string]
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "lore", decoded.V["kind"])
}
