module github.com/plexusone/omni-anthropic

go 1.26.0

require (
	github.com/anthropics/anthropic-sdk-go v1.48.0
	github.com/plexusone/omnillm-core v0.16.0
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.2.0 // indirect
	github.com/grokify/mogo v0.74.5 // indirect
	github.com/grokify/sogo v0.15.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/standard-webhooks/standard-webhooks/libraries v0.0.1 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
	golang.org/x/sync v0.20.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Exclude jsonschema v0.14.0+ due to incompatibility with anthropic-sdk-go.
// anthropic-sdk-go uses wk8/go-ordered-map but jsonschema v0.14.0 switched to pb33f/ordered-map.
exclude github.com/invopop/jsonschema v0.14.0
