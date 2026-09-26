package forkdocs

import _ "embed"

//go:embed swagger.yaml
var openAPISpec []byte

// OpenAPISpec returns the embedded OpenAPI (Swagger 2.0) specification of the
// API served by this instance. All described paths are relative to the
// basePath /api.
func OpenAPISpec() []byte {
	return openAPISpec
}
