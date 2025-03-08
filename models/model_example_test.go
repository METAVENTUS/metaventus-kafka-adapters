package models

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/hamba/avro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestModelExample_GetSchema(t *testing.T) {
	m := ModelExample{
		ID:    uuid.NewString(),
		Email: "john@doe.com",
		Name:  "john doe",
	}

	buf := new(bytes.Buffer)
	encoder, err := avro.NewEncoder(m.GetSchema(), buf)
	require.NoError(t, err)

	err = encoder.Encode(m)
	require.NoError(t, err)

	var event ModelExample
	decoder, err := avro.NewDecoder(event.GetSchema(), buf)
	require.NoError(t, err)

	err = decoder.Decode(&event)
	require.NoError(t, err)

	assert.Equal(t, m, event)
}
