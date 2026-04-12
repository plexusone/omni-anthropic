# OmniLLM Provider for Anthropic

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omnillm-anthropic/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omnillm-anthropic
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omnillm-anthropic
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnillm-anthropic
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnillm-anthropic
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnillm-anthropic/blob/master/LICENSE

[OmniLLM](https://github.com/plexusone/omnillm-core) thick provider for Anthropic using the official [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) SDK.

## Installation

```bash
go get github.com/plexusone/omnillm-anthropic
```

## Usage

### Direct Usage

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/omnillm-anthropic"
    "github.com/plexusone/omnillm-core/provider"
)

func main() {
    p, err := anthropic.New(anthropic.Config{
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()

    resp, err := p.CreateChatCompletion(context.Background(), &provider.ChatCompletionRequest{
        Model: "claude-sonnet-4-20250514",
        Messages: []provider.Message{
            {Role: provider.RoleUser, Content: "Hello!"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Println(resp.Choices[0].Message.Content)
}
```

### Auto-Registration with OmniLLM

Importing this package automatically registers Anthropic as a thick provider:

```go
package main

import (
    "context"
    "log"

    omnillm "github.com/plexusone/omnillm-core"
    _ "github.com/plexusone/omnillm-anthropic" // Auto-registers Anthropic thick provider
)

func main() {
    client, err := omnillm.NewClient(omnillm.ClientConfig{
        Provider: omnillm.ProviderNameAnthropic,
        APIKey:   "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Use client as usual...
}
```

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `APIKey` | string | Yes | Anthropic API key |
| `BaseURL` | string | No | Custom API endpoint |

## Features

- Chat completions (non-streaming and streaming)
- Tool/function calling support
- System prompts and multi-turn conversations
- Automatic priority registration (overrides thin provider)

## Why a Thick Provider?

Thick providers use official SDKs, providing:

- Full API coverage and latest features
- Automatic retries and error handling
- SDK-managed authentication

The trade-off is additional dependencies. Import this package only if you need Anthropic support.

## Testing

```bash
go test -v ./...
```

## License

MIT
