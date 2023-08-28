package orchestrator

import (
	"bytes"

	artifact "github.com/DeepSourceCorp/artifacts/types"
	"github.com/pelletier/go-toml"
)

type MarvinAnalysisConfig struct {
	*artifact.MarvinAnalysisConfig
}

func NewMarvinAnalysisConfig(run *artifact.AnalysisRun, check artifact.Check) *MarvinAnalysisConfig {
	return &MarvinAnalysisConfig{
		MarvinAnalysisConfig: &artifact.MarvinAnalysisConfig{
			RunID:                      run.RunID,
			CheckSeq:                   check.CheckSeq,
			AnalyzerShortcode:          check.AnalyzerMeta.Shortcode,
			AnalyzerCommand:            check.AnalyzerMeta.Command,
			BaseOID:                    run.VCSMeta.BaseOID,
			CheckoutOID:                run.VCSMeta.CheckoutOID,
			IsForDefaultAnalysisBranch: run.VCSMeta.IsForDefaultAnalysisBranch,
			DSConfigUpdated:            run.DSConfigUpdated,
			Processors:                 check.Processors,
			DiffMetaCommits:            check.DiffMetaCommits,
		},
	}
}

func (c *MarvinAnalysisConfig) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
