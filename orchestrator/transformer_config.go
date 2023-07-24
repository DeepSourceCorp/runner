package orchestrator

import (
	"bytes"

	artifact "github.com/deepsourcelabs/artifacts/types"
	"github.com/pelletier/go-toml"
)

type TransformerMarvinConfig struct {
	*artifact.MarvinTransformerConfig
}

func NewTransformerMarvinConfig(run *artifact.TransformerRun) *TransformerMarvinConfig {
	return &TransformerMarvinConfig{
		&artifact.MarvinTransformerConfig{
			RunID:              run.RunID,
			BaseOID:            run.VCSMeta.BaseOID,
			CheckoutOID:        run.VCSMeta.CheckoutOID,
			TransformerCommand: run.Transformer.Command,
			TransformerTools:   run.Transformer.Tools,
			DSConfigUpdated:    run.DSConfigUpdated,
			PatchCommit:        run.PatchCommit,
		},
	}
}

func (c *TransformerMarvinConfig) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
