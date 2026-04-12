package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	core "github.com/plexusone/omnillm-core"
)

func init() {
	// Register Anthropic as a thick provider (priority 10, overrides thin provider)
	core.RegisterProvider(core.ProviderNameAnthropic, newProviderFromConfig, core.PriorityThick)
}

// newProviderFromConfig creates a new Anthropic provider from omnillm config (for registry).
func newProviderFromConfig(config core.ProviderConfig) (core.Provider, error) {
	return New(Config{
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
}

// Config holds configuration for the Anthropic provider.
type Config struct {
	// APIKey is the Anthropic API key (required).
	APIKey string

	// BaseURL is an optional custom API endpoint.
	BaseURL string
}

// Provider implements core.Provider using the official Anthropic SDK.
type Provider struct {
	client anthropic.Client
	config Config
}

// Ensure Provider implements core.Provider at compile time.
var _ core.Provider = (*Provider)(nil)

// New creates a new Anthropic provider with the given configuration.
func New(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, core.ErrInvalidAPIKey
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	return &Provider{
		client: client,
		config: cfg,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "anthropic"
}

// Capabilities returns the provider's supported features.
func (p *Provider) Capabilities() core.Capabilities {
	return core.Capabilities{
		Tools:             true,
		Streaming:         true,
		Vision:            true,
		JSON:              false, // Anthropic doesn't have a JSON mode like OpenAI
		SystemRole:        true,  // Handled via separate system field
		MaxContextWindow:  200000, // Claude 3 models
		SupportsMaxTokens: true,
	}
}

// Close releases resources held by the provider.
func (p *Provider) Close() error {
	// The Anthropic SDK doesn't require explicit cleanup
	return nil
}

// CreateChatCompletion sends a chat completion request and returns the response.
func (p *Provider) CreateChatCompletion(ctx context.Context, req *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	params := p.buildParams(req)

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, p.wrapError(err)
	}

	return p.convertResponse(resp), nil
}

// CreateChatCompletionStream creates a streaming chat completion.
func (p *Provider) CreateChatCompletionStream(ctx context.Context, req *core.ChatCompletionRequest) (core.ChatCompletionStream, error) {
	params := p.buildParams(req)

	stream := p.client.Messages.NewStreaming(ctx, params)

	return &streamAdapter{stream: stream}, nil
}

// buildParams converts a core request to Anthropic SDK params.
func (p *Provider) buildParams(req *core.ChatCompletionRequest) anthropic.MessageNewParams {
	// Default max tokens if not provided
	maxTokens := int64(4096)
	if req.MaxTokens != nil {
		maxTokens = int64(*req.MaxTokens)
	}

	params := anthropic.MessageNewParams{
		Model:     req.Model,
		MaxTokens: maxTokens,
		Messages:  p.convertMessages(req.Messages),
	}

	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
	}
	if req.TopK != nil {
		params.TopK = param.NewOpt(int64(*req.TopK))
	}
	if len(req.Stop) > 0 {
		params.StopSequences = req.Stop
	}

	// Extract system message and convert to system param
	for _, msg := range req.Messages {
		if msg.Role == core.RoleSystem {
			params.System = []anthropic.TextBlockParam{
				{Text: msg.Content},
			}
			break
		}
	}

	// Tools
	if len(req.Tools) > 0 {
		params.Tools = p.convertTools(req.Tools)
	}

	// Tool choice
	if req.ToolChoice != nil {
		params.ToolChoice = p.convertToolChoice(req.ToolChoice)
	}

	return params
}

// convertMessages converts core messages to Anthropic message params.
func (p *Provider) convertMessages(messages []core.Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case core.RoleSystem:
			// System messages are handled separately in Anthropic
			continue

		case core.RoleUser:
			result = append(result, anthropic.NewUserMessage(
				anthropic.NewTextBlock(msg.Content),
			))

		case core.RoleAssistant:
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.ToolCalls)+1)
				if msg.Content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
				}
				for _, tc := range msg.ToolCalls {
					// Parse the arguments JSON
					var input any
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
						input = map[string]any{"raw": tc.Function.Arguments}
					}
					blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Function.Name))
				}
				result = append(result, anthropic.NewAssistantMessage(blocks...))
			} else {
				result = append(result, anthropic.NewAssistantMessage(
					anthropic.NewTextBlock(msg.Content),
				))
			}

		case core.RoleTool:
			// Tool results in Anthropic are user messages with tool_result content blocks
			toolCallID := ""
			if msg.ToolCallID != nil {
				toolCallID = *msg.ToolCallID
			}
			// Check if we need to append to the last user message or create new one
			if len(result) > 0 && result[len(result)-1].Role == anthropic.MessageParamRoleUser {
				// Append to existing user message
				result[len(result)-1].Content = append(
					result[len(result)-1].Content,
					anthropic.NewToolResultBlock(toolCallID, msg.Content, false),
				)
			} else {
				// Create new user message with tool result
				result = append(result, anthropic.NewUserMessage(
					anthropic.NewToolResultBlock(toolCallID, msg.Content, false),
				))
			}
		}
	}

	return result
}

// convertTools converts core tools to Anthropic tool params.
func (p *Provider) convertTools(tools []core.Tool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, 0, len(tools))

	for _, tool := range tools {
		// Convert parameters to ToolInputSchemaParam
		var inputSchema anthropic.ToolInputSchemaParam
		if tool.Function.Parameters != nil {
			// Parameters should be a JSON Schema object
			if params, ok := tool.Function.Parameters.(map[string]any); ok {
				// Extract properties if present
				if props, ok := params["properties"]; ok {
					inputSchema.Properties = props
				} else {
					inputSchema.Properties = params
				}
				// Extract required if present
				if req, ok := params["required"].([]any); ok {
					for _, r := range req {
						if s, ok := r.(string); ok {
							inputSchema.Required = append(inputSchema.Required, s)
						}
					}
				}
			}
		}

		result = append(result, anthropic.ToolUnionParamOfTool(inputSchema, tool.Function.Name))
		// Set description on the last added tool
		if len(result) > 0 && result[len(result)-1].OfTool != nil {
			result[len(result)-1].OfTool.Description = param.NewOpt(tool.Function.Description)
		}
	}

	return result
}

// convertToolChoice converts core tool choice to Anthropic format.
func (p *Provider) convertToolChoice(choice any) anthropic.ToolChoiceUnionParam {
	switch v := choice.(type) {
	case string:
		switch v {
		case "auto":
			return anthropic.ToolChoiceUnionParam{
				OfAuto: &anthropic.ToolChoiceAutoParam{},
			}
		case "none":
			return anthropic.ToolChoiceUnionParam{
				OfNone: &anthropic.ToolChoiceNoneParam{},
			}
		case "required":
			return anthropic.ToolChoiceUnionParam{
				OfAny: &anthropic.ToolChoiceAnyParam{},
			}
		}
	case map[string]any:
		// Specific function choice
		if fn, ok := v["function"].(map[string]any); ok {
			if name, ok := fn["name"].(string); ok {
				return anthropic.ToolChoiceParamOfTool(name)
			}
		}
	}
	return anthropic.ToolChoiceUnionParam{
		OfAuto: &anthropic.ToolChoiceAutoParam{},
	}
}

// convertResponse converts an Anthropic response to core format.
func (p *Provider) convertResponse(resp *anthropic.Message) *core.ChatCompletionResponse {
	result := &core.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   string(resp.Model),
		Usage: core.Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}

	// Convert content blocks
	var content string
	var toolCalls []core.ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			// Convert input to JSON string
			inputStr := "{}"
			if block.Input != nil {
				if data, err := json.Marshal(block.Input); err == nil {
					inputStr = string(data)
				}
			}
			toolCalls = append(toolCalls, core.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: core.ToolFunction{
					Name:      block.Name,
					Arguments: inputStr,
				},
			})
		}
	}

	// Determine finish reason
	finishReason := string(resp.StopReason)
	if len(toolCalls) > 0 && finishReason == "end_turn" {
		finishReason = "tool_calls"
	}

	result.Choices = []core.ChatCompletionChoice{
		{
			Index: 0,
			Message: core.Message{
				Role:      core.RoleAssistant,
				Content:   content,
				ToolCalls: toolCalls,
			},
			FinishReason: &finishReason,
		},
	}

	// Preserve Anthropic-specific metadata
	result.ProviderMetadata = map[string]any{
		"anthropic_stop_reason": resp.StopReason,
	}

	return result
}

// wrapError converts Anthropic SDK errors to core errors.
func (p *Provider) wrapError(err error) error {
	if err == nil {
		return nil
	}

	return core.NewAPIError("anthropic", 0, "", err.Error())
}

// streamAdapter wraps an Anthropic stream to implement core.ChatCompletionStream.
type streamAdapter struct {
	stream    *ssestream.Stream[anthropic.MessageStreamEventUnion]
	messageID string
	model     string
}

// Recv receives the next chunk from the stream.
func (s *streamAdapter) Recv() (*core.ChatCompletionChunk, error) {
	if !s.stream.Next() {
		err := s.stream.Err()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	event := s.stream.Current()
	return s.convertEvent(event), nil
}

// Close closes the stream.
func (s *streamAdapter) Close() error {
	return s.stream.Close()
}

// convertEvent converts an Anthropic streaming event to core format.
func (s *streamAdapter) convertEvent(event anthropic.MessageStreamEventUnion) *core.ChatCompletionChunk {
	result := &core.ChatCompletionChunk{
		ID:      s.messageID,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   s.model,
	}

	switch event.Type {
	case "message_start":
		// Store message metadata for future chunks
		if msg := event.AsMessageStart(); msg.Message.ID != "" {
			s.messageID = msg.Message.ID
			s.model = string(msg.Message.Model)
			result.ID = s.messageID
			result.Model = s.model
		}
		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
		}

	case "content_block_start":
		startEvent := event.AsContentBlockStart()
		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
			"anthropic_index":      startEvent.Index,
		}

		// Check if this is a tool_use block
		if startEvent.ContentBlock.Type == "tool_use" {
			result.Choices = []core.ChatCompletionChoice{
				{
					Index: 0,
					Delta: &core.Message{
						Role: core.RoleAssistant,
						ToolCalls: []core.ToolCall{
							{
								ID:   startEvent.ContentBlock.ID,
								Type: "function",
								Function: core.ToolFunction{
									Name: startEvent.ContentBlock.Name,
								},
							},
						},
					},
				},
			}
		}

	case "content_block_delta":
		deltaEvent := event.AsContentBlockDelta()
		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
			"anthropic_index":      deltaEvent.Index,
		}

		// Check delta type
		if deltaEvent.Delta.Type == "text_delta" {
			result.Choices = []core.ChatCompletionChoice{
				{
					Index: 0,
					Delta: &core.Message{
						Role:    core.RoleAssistant,
						Content: deltaEvent.Delta.Text,
					},
				},
			}
		} else if deltaEvent.Delta.Type == "input_json_delta" {
			// Tool input streaming
			result.ProviderMetadata["tool_input_delta"] = deltaEvent.Delta.PartialJSON
		}

	case "message_delta":
		deltaEvent := event.AsMessageDelta()
		var finishReason *string
		if deltaEvent.Delta.StopReason != "" {
			reason := string(deltaEvent.Delta.StopReason)
			finishReason = &reason
		}

		result.Choices = []core.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: finishReason,
			},
		}

		if deltaEvent.Usage.OutputTokens > 0 {
			result.Usage = &core.Usage{
				CompletionTokens: int(deltaEvent.Usage.OutputTokens),
			}
		}

		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
		}

	case "message_stop":
		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
		}

	default:
		result.ProviderMetadata = map[string]any{
			"anthropic_event_type": event.Type,
		}
	}

	return result
}
