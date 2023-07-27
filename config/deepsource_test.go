package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDeepSource_UnmarshalYAML(t *testing.T) {
	input := `
host: https://deepsource.io
publicKey: |
  -----BEGIN PUBLIC KEY-----
  MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAq9uoCTPIYLfIuUj02YOC
  KPjoqCCbUIO7YSXg1DASwd9snFwffCDY+sxbXl//z5Lqg/JQhDcW7DkA1QmFKtB/
  mpMuE1RSlx4n96eoEE15FP2Kqd/RFptA3TDHqziNK+ydczLMmMP+o70lFhwGWKeL
  BGoVpp/0/GQvpiWWe/PRVbpy8gm1rWJeA5hJJLgNaJRWnF3+ocihIwWdwTsPnZCR
  3w1KQjZp2+Y9NBL92W+5jwrIaMtvzV+f3t/imQ2Rgy/c21pDbGKA9Z/ddLFVxnoD
  y1PUzFM+RUKElT8GsX/Y+LEEzTzqdCJYm/MfKVjn7OyMtwN112TQ5ZvZkIzYVLzf
  uwIDAQAB
  -----END PUBLIC KEY-----
`
	var d DeepSource
	err := yaml.Unmarshal([]byte(input), &d)
	require.NoError(t, err)
	assert.Equal(t, "https://deepsource.io", d.Host.String())
	assert.NotNil(t, d.PublicKey)
}
