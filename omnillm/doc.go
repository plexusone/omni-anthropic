// Package anthropic provides an OmniLLM adapter for Anthropic's Claude API using the official anthropic-sdk-go.
//
// This package implements the [core.Provider] interface from omnillm-core,
// wrapping the official Anthropic Go SDK (github.com/anthropics/anthropic-sdk-go).
//
// # Features
//
// The adapter supports all major Anthropic features:
//   - Chat completions (Claude 3 Opus, Sonnet, Haiku, etc.)
//   - Streaming responses
//   - Tool/function calling
//   - Vision (image inputs)
//   - System prompts
//
// # Basic Usage
//
//	import (
//	    core "github.com/plexusone/omnillm-core"
//	    "github.com/plexusone/omni-anthropic/omnillm"
//	)
//
//	func main() {
//	    provider, err := anthropic.New(anthropic.Config{
//	        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer provider.Close()
//
//	    resp, err := provider.CreateChatCompletion(ctx, &core.ChatCompletionRequest{
//	        Model: "claude-3-opus-20240229",
//	        Messages: []core.Message{
//	            {Role: core.RoleUser, Content: "Hello!"},
//	        },
//	    })
//	}
//
// # Configuration
//
// The [Config] struct supports:
//   - APIKey: Your Anthropic API key (required)
//   - BaseURL: Custom API endpoint (optional)
//
// # Streaming
//
//	stream, err := provider.CreateChatCompletionStream(ctx, &core.ChatCompletionRequest{
//	    Model: "claude-3-sonnet-20240229",
//	    Messages: []core.Message{
//	        {Role: core.RoleUser, Content: "Tell me a story"},
//	    },
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stream.Close()
//
//	for {
//	    chunk, err := stream.Recv()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    // Process chunk
//	}
package anthropic
