# Release Notes - v0.1.0

Initial release of omnillm-anthropic, a thick provider for [OmniLLM](https://github.com/plexusone/omnillm-core) using the official [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) SDK.

## Features

- **Chat Completion** - Full support for Anthropic Messages API
- **Streaming** - Real-time streaming responses
- **Tool Calling** - Function/tool calling support
- **System Messages** - System prompt support
- **Auto-Registration** - Automatically registers as thick provider via `init()`

## Feature Support

| Feature | Supported |
|---------|-----------|
| Chat Completion | Yes |
| Streaming | Yes |
| Tool Calling | Yes |
| System Messages | Yes |
| JSON Mode | No |

## Installation

```bash
go get github.com/plexusone/omnillm-anthropic@v0.1.0
```

## Usage

```go
import (
    omnillm "github.com/plexusone/omnillm-core"
    _ "github.com/plexusone/omnillm-anthropic" // Auto-registers thick provider
)

client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameAnthropic,
    APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
})
```

## Dependencies

- `github.com/anthropics/anthropic-sdk-go` v1.35.0
- `github.com/plexusone/omnillm-core` v0.15.0

## Testing

Includes conformance tests using the `providertest` framework from omnillm-core.

```bash
# Run unit tests
go test -v ./...

# Run integration tests (requires API key)
ANTHROPIC_API_KEY=your-key go test -v ./...
```
