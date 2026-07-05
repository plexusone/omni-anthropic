package anthropic

import (
	"context"
	"os"
	"testing"
	"time"

	core "github.com/plexusone/omnillm-core/provider"
)

func TestConvertThinkingConfig(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name      string
		thinkType string
		budget    *int64
		wantType  string // "disabled", "adaptive", "enabled"
	}{
		{"disabled", core.ThinkingTypeDisabled, nil, "disabled"},
		{"adaptive", core.ThinkingTypeAdaptive, nil, "adaptive"},
		{"enabled without budget", core.ThinkingTypeEnabled, nil, "enabled"},
		{"enabled with budget", core.ThinkingTypeEnabled, ptr(int64(4096)), "enabled"},
		{"unknown defaults to adaptive", "unknown", nil, "adaptive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &core.ThinkingConfig{
				Type:         tt.thinkType,
				BudgetTokens: tt.budget,
			}

			got := p.convertThinkingConfig(cfg, nil)

			switch tt.wantType {
			case "disabled":
				if got.OfDisabled == nil {
					t.Error("expected OfDisabled to be set")
				}
			case "adaptive":
				if got.OfAdaptive == nil {
					t.Error("expected OfAdaptive to be set")
				}
			case "enabled":
				if got.OfEnabled == nil {
					t.Error("expected OfEnabled to be set")
				}
			}
		})
	}
}

func TestMapReasoningEffortToThinking(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name      string
		effort    string
		maxTokens *int
		wantType  string // "disabled", "adaptive", "enabled"
	}{
		{"none maps to disabled", core.ReasoningEffortNone, nil, "disabled"},
		{"low maps to adaptive", core.ReasoningEffortLow, nil, "adaptive"},
		{"medium maps to adaptive", core.ReasoningEffortMedium, nil, "adaptive"},
		{"high maps to enabled", core.ReasoningEffortHigh, nil, "enabled"},
		{"unknown maps to adaptive", "unknown", nil, "adaptive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.mapReasoningEffortToThinking(tt.effort, tt.maxTokens)

			switch tt.wantType {
			case "disabled":
				if got.OfDisabled == nil {
					t.Error("expected OfDisabled to be set")
				}
			case "adaptive":
				if got.OfAdaptive == nil {
					t.Error("expected OfAdaptive to be set")
				}
			case "enabled":
				if got.OfEnabled == nil {
					t.Error("expected OfEnabled to be set")
				}
			}
		})
	}
}

func TestMapReasoningEffortToThinking_BudgetCalculation(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name          string
		maxTokens     *int
		wantMinBudget int64
	}{
		{"nil maxTokens uses default 8192", nil, 8192},
		{"small maxTokens uses default 8192", ptr(1000), 8192},
		{"large maxTokens uses 25%", ptr(100000), 25000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.mapReasoningEffortToThinking(core.ReasoningEffortHigh, tt.maxTokens)

			if got.OfEnabled == nil {
				t.Fatal("expected OfEnabled to be set")
			}
			if got.OfEnabled.BudgetTokens < tt.wantMinBudget {
				t.Errorf("BudgetTokens = %d, want >= %d", got.OfEnabled.BudgetTokens, tt.wantMinBudget)
			}
		})
	}
}

func TestConvertThinkingConfig_EnabledBudget(t *testing.T) {
	p := &Provider{}

	// Test with explicit budget
	budget := int64(4096)
	cfg := &core.ThinkingConfig{
		Type:         core.ThinkingTypeEnabled,
		BudgetTokens: &budget,
	}

	got := p.convertThinkingConfig(cfg, nil)

	if got.OfEnabled == nil {
		t.Fatal("expected OfEnabled to be set")
	}
	if got.OfEnabled.BudgetTokens != budget {
		t.Errorf("BudgetTokens = %d, want %d", got.OfEnabled.BudgetTokens, budget)
	}
}

func TestConvertThinkingConfig_EnabledDefaultBudget(t *testing.T) {
	p := &Provider{}

	// Test without explicit budget - should use default 8192
	cfg := &core.ThinkingConfig{
		Type: core.ThinkingTypeEnabled,
	}

	got := p.convertThinkingConfig(cfg, nil)

	if got.OfEnabled == nil {
		t.Fatal("expected OfEnabled to be set")
	}
	if got.OfEnabled.BudgetTokens != 8192 {
		t.Errorf("BudgetTokens = %d, want 8192 (default)", got.OfEnabled.BudgetTokens)
	}
}

// ptr returns a pointer to the value
func ptr[T any](v T) *T {
	return &v
}

// Integration tests - require ANTHROPIC_API_KEY

func TestIntegration_ThinkingConfig_Adaptive(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	p, err := New(Config{APIKey: apiKey})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test with adaptive thinking
	maxTokens := 200
	req := &core.ChatCompletionRequest{
		Model:     "claude-sonnet-4-6",
		MaxTokens: &maxTokens,
		Thinking: &core.ThinkingConfig{
			Type: core.ThinkingTypeAdaptive,
		},
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "What is 2+2? Reply with just the number."},
		},
	}

	resp, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("CreateChatCompletion() returned no choices")
	}

	t.Logf("Response with thinking=adaptive: %s", resp.Choices[0].Message.Content)
}

func TestIntegration_ThinkingConfig_Enabled(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	p, err := New(Config{APIKey: apiKey})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Test with enabled thinking and explicit budget
	maxTokens := 16000
	budget := int64(10000)
	req := &core.ChatCompletionRequest{
		Model:     "claude-sonnet-4-6",
		MaxTokens: &maxTokens,
		Thinking: &core.ThinkingConfig{
			Type:         core.ThinkingTypeEnabled,
			BudgetTokens: &budget,
		},
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "What is 15 * 17? Think step by step, then give the answer."},
		},
	}

	resp, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("CreateChatCompletion() returned no choices")
	}

	t.Logf("Response with thinking=enabled budget=%d: %s", budget, resp.Choices[0].Message.Content)
}

func TestIntegration_ReasoningEffort_MapsToThinking(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	p, err := New(Config{APIKey: apiKey})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test with ReasoningEffort which should map to Anthropic thinking
	effort := core.ReasoningEffortMedium
	maxTokens := 200
	req := &core.ChatCompletionRequest{
		Model:           "claude-sonnet-4-6",
		MaxTokens:       &maxTokens,
		ReasoningEffort: &effort,
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "What is 2+2? Reply with just the number."},
		},
	}

	resp, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() error: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("CreateChatCompletion() returned no choices")
	}

	t.Logf("Response with reasoning_effort=%s (mapped to adaptive): %s", effort, resp.Choices[0].Message.Content)
}

func TestIntegration_ThinkingConfig_Streaming(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	p, err := New(Config{APIKey: apiKey})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	maxTokens := 200
	req := &core.ChatCompletionRequest{
		Model:     "claude-sonnet-4-6",
		MaxTokens: &maxTokens,
		Thinking: &core.ThinkingConfig{
			Type: core.ThinkingTypeAdaptive,
		},
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "What is 2+2? Reply with just the number."},
		},
	}

	stream, err := p.CreateChatCompletionStream(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletionStream() error: %v", err)
	}
	defer stream.Close()

	var content string
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			content += chunk.Choices[0].Delta.Content
		}
	}

	if content == "" {
		t.Error("Streaming returned no content")
	}

	t.Logf("Streaming response with thinking=adaptive: %s", content)
}
