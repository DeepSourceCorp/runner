package orchestrator

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	artifact "github.com/DeepSourceCorp/artifacts/types"
	"github.com/klauspost/compress/zstd"
	"github.com/pelletier/go-toml"
)

type AutofixConfig struct {
	*artifact.MarvinAutofixConfig
}

func NewAutofixConfig(run *artifact.AutofixRun) (*AutofixConfig, error) {
	jsonIssues, err := json.Marshal(run.Autofixer.Autofixes)
	if err != nil {
		return nil, err
	}
	return &AutofixConfig{
		MarvinAutofixConfig: &artifact.MarvinAutofixConfig{
			RunID:             run.RunID,
			AnalyzerShortcode: run.Autofixer.AutofixMeta.Shortcode,
			AutofixerCommand:  run.Autofixer.AutofixMeta.Command,
			CheckoutOID:       run.VCSMeta.CheckoutOID,
			AutofixIssues:     string(jsonIssues),
		},
	}, nil
}

func (c *AutofixConfig) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}

	var compressed []byte
	w, err := zstd.NewWriter(bytes.NewBuffer(compressed), zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(compressed)))
	base64.StdEncoding.Encode(encoded, compressed)
	return encoded, nil
}
