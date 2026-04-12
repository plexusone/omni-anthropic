# OmniLLM Provider for Anthropic

[![Go Reference][docs-godoc-svg]][docs-godoc-url]
[![Go Report Card][goreport-svg]][goreport-url]

 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnillm-anthropic
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnillm-anthropic
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omnillm-anthropic
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omnillm-anthropic

Thick provider for [OmniLLM](https://github.com/plexusone/omnillm-core) using the official [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) SDK.

## Installation

```bash
go get github.com/plexusone/omnillm-anthropic
```

## Quick Start

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

## Feature Support

| Feature | Supported |
|---------|-----------|
| Chat Completion | Yes |
| Streaming | Yes |
| Tool Calling | Yes |
| System Messages | Yes |
| JSON Mode | No |

## Configuration

| Field | Required | Description |
|-------|----------|-------------|
| `APIKey` | Yes | Anthropic API key |
| `BaseURL` | No | Custom endpoint |

## Documentation

See [OmniLLM Core](https://github.com/plexusone/omnillm-core) for full API documentation.

## License

MIT
