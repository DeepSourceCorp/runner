package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestKubernetes_UnmarshalYAML(t *testing.T) {
	input := `
namespace: default
nodeSelector:
  foo: bar`
	var k Kubernetes
	err := yaml.Unmarshal([]byte(input), &k)
	_ = err
	require.NoError(t, err)
	assert.Equal(t, "default", k.Namespace)
	assert.Equal(t, map[string]string{"foo": "bar"}, k.NodeSelector)
}
