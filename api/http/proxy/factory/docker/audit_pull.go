package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/portainer/portainer/api/internal/audit"
)

const pullBodyTailSize = 32 * 1024

// auditImagePull wraps the Docker image creation response so the streamed
// pull/import result is parsed and recorded as a user activity log entry once
// the body is consumed. The body itself is passed through untouched.
func (transport *Transport) auditImagePull(request *http.Request, response *http.Response) {
	query := request.URL.Query()

	image := query.Get("fromImage")
	action := "image_pull"
	if image == "" {
		image = query.Get("fromSrc")
		action = "image_import"
	}
	if image == "" {
		return
	}
	if tag := query.Get("tag"); tag != "" {
		image = image + ":" + tag
	}

	onDone := func(tail string, aborted bool) {
		payload := parsePullStream(tail)
		payload["image"] = image
		if aborted {
			payload["error"] = "pull stream was not fully consumed"
		}

		if response.StatusCode >= http.StatusBadRequest && payload["error"] == nil {
			payload["error"] = errorFromTail(tail, response.StatusCode)
		}

		success := !aborted && payload["error"] == nil
		auditor := audit.NewAuditor(transport.dataStore)
		auditor.Record(request, transport.endpoint.ID, action, payload, success)
	}

	response.Body = newPullAuditBody(response.Body, onDone)
}

type pullAuditBody struct {
	io.ReadCloser
	tail   []byte
	onDone func(tail string, aborted bool)
	done   bool
}

func newPullAuditBody(body io.ReadCloser, onDone func(tail string, aborted bool)) io.ReadCloser {
	return &pullAuditBody{ReadCloser: body, onDone: onDone}
}

func (b *pullAuditBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.tail = append(b.tail, p[:n]...)
		if len(b.tail) > pullBodyTailSize {
			b.tail = b.tail[len(b.tail)-pullBodyTailSize:]
		}
	}
	if err == io.EOF && !b.done {
		b.done = true
		b.onDone(string(b.tail), false)
	}
	return n, err
}

func (b *pullAuditBody) Close() error {
	if !b.done {
		b.done = true
		b.onDone(string(b.tail), true)
	}
	return b.ReadCloser.Close()
}

// errorFromTail extracts the failure reason from a non-stream error response
// (Docker returns a JSON body with a "message" field on failed API requests).
func errorFromTail(tail string, statusCode int) string {
	for _, line := range strings.Split(tail, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var message struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &message); err == nil && message.Message != "" {
			return message.Message
		}
	}

	tail = strings.TrimSpace(tail)
	if tail != "" {
		if len(tail) > 512 {
			tail = tail[:512]
		}
		return tail
	}

	return fmt.Sprintf("docker API responded with status %d", statusCode)
}

// parsePullStream extracts the result from the tail of a Docker JSON message
// stream (pull/import progress lines).
func parsePullStream(tail string) map[string]any {
	payload := map[string]any{}

	for _, line := range strings.Split(tail, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var message struct {
			Status string `json:"status"`
			Error  string `json:"error"`
			Aux    *struct {
				ID string `json:"ID"`
			} `json:"aux"`
		}
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			continue
		}

		switch {
		case message.Error != "":
			payload["error"] = message.Error
		case message.Aux != nil && message.Aux.ID != "":
			payload["digest"] = message.Aux.ID
		case strings.Contains(message.Status, "Downloaded"):
			payload["downloaded"] = true
		case strings.Contains(message.Status, "up to date") || strings.Contains(message.Status, "up-to-date"):
			payload["downloaded"] = false
		}
	}

	return payload
}
