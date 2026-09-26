package docker

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parsePullStream(t *testing.T) {
	t.Run("detects downloaded image with digest", func(t *testing.T) {
		tail := `{"status":"Pulling from library/nginx","id":"latest"}
{"status":"Downloading","id":"a1b2c3","progressDetail":{"current":100,"total":200},"progress":"[===>  ]"}
{"status":"Extracting","id":"a1b2c3","progressDetail":{"current":200,"total":200}}
{"status":"Downloaded newer image for nginx:latest"}
{"aux":{"ID":"sha256:abc123"}}
{"status":"Status: Downloaded newer image for nginx:latest"}
`
		payload := parsePullStream(tail)

		assert.Equal(t, true, payload["downloaded"])
		assert.Equal(t, "sha256:abc123", payload["digest"])
		assert.NotContains(t, payload, "error")
	})

	t.Run("detects up to date image", func(t *testing.T) {
		payload := parsePullStream(`{"status":"Status: Image is up to date for nginx:latest"}`)

		assert.Equal(t, false, payload["downloaded"])
	})

	t.Run("extracts error details", func(t *testing.T) {
		payload := parsePullStream(`{"status":"Pulling from library/nginx"}
{"errorDetail":{"message":"pull access denied"},"error":"pull access denied for repo"}
`)

		assert.Equal(t, "pull access denied for repo", payload["error"])
	})

	t.Run("ignores non json garbage", func(t *testing.T) {
		payload := parsePullStream("not json\n\npartial line")

		assert.Empty(t, payload)
	})
}

func Test_errorFromTail(t *testing.T) {
	t.Run("extracts docker error message", func(t *testing.T) {
		err := errorFromTail(`{"message":"no such image"}`, http.StatusNotFound)

		assert.Equal(t, "no such image", err)
	})

	t.Run("falls back to raw body then status", func(t *testing.T) {
		assert.Equal(t, "plain text failure", errorFromTail("plain text failure", http.StatusNotFound))
		assert.Equal(t, "docker API responded with status 500", errorFromTail("", http.StatusInternalServerError))
	})
}
