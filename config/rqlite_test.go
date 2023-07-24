package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRQLite_UnmarshalYAML(t *testing.T) {
	input := `
host: localhost
port: 4001`
	var r RQLite
	err := yaml.Unmarshal([]byte(input), &r)
	require.NoError(t, err)
	assert.Equal(t, "localhost", r.Host)
	assert.Equal(t, 4001, r.Port)
}
