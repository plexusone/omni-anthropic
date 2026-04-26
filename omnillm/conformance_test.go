package anthropic

import (
	"os"
	"testing"

	"github.com/plexusone/omnillm-core/provider/providertest"
)

func TestConformance(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	p, err := New(Config{APIKey: apiKey})
	if err != nil && apiKey != "" {
		t.Fatalf("New() error: %v", err)
	}
	if p == nil {
		// Create with empty key for interface tests only
		p = &Provider{}
	}

	providertest.RunAll(t, providertest.Config{
		Provider:        p,
		SkipIntegration: apiKey == "",
		TestModel:       "claude-3-haiku-20240307",
	})
}
