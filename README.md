# OmniLLM Provider for Anthropic

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
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
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomnillm-anthropic
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omnillm-anthropic
 [repo-url]: https://github.com/plexusone/omnillm-anthropic
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnillm-anthropic/blob/master/LICENSE

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
