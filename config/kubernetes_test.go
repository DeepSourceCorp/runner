package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestKubernetes_UnmarshalYAML(t *testing.T) {
	t.Setenv("TASK_IMAGE_PULL_SECRET_NAME", "default")
	t.Setenv("TASK_IMAGE_REGISTRY_URL", "example.com")
	t.Setenv("TASK_NAMESPACE", "default")
	t.Setenv("TASK_NODE_SELECTOR", "bar: foo")
	input := `
namespace: analysis
nodeSelector:
  foo: bar`
	var k Kubernetes
	err := yaml.Unmarshal([]byte(input), &k)
	_ = err
	require.NoError(t, err)
	assert.Equal(t, "analysis", k.Namespace)
	assert.Equal(t, map[string]string{"foo": "bar"}, k.NodeSelector)
	assert.Equal(t, "default", k.ImageRegistry.PullSecretName)

	u, _ := url.Parse("example.com")
	assert.Equal(t, *u, k.ImageRegistry.RegistryUrl)
}
