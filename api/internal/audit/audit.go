package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Auditor records user activity log entries.
type Auditor struct {
	dataStore dataservices.DataStore
}

// NewAuditor creates a new Auditor.
func NewAuditor(dataStore dataservices.DataStore) *Auditor {
	return &Auditor{dataStore: dataStore}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// DataStoreAccessor resolves the datastore lazily. Handlers receive their
// datastore after NewHandler returns, so route wrappers must not capture the
// value at registration time.
type DataStoreAccessor func() dataservices.DataStore

// Middleware wraps an API handler so that every call is recorded as a user
// activity log entry. The request body is replayed so the wrapped handler
// keeps working; the selected JSON body fields and the route variables are
// copied into the log payload. A non-nil handler error is treated as a
// failed action and its message is recorded in the payload.
func Middleware(dataStore DataStoreAccessor, action string, payloadKeys ...string) func(next httperror.LoggerHandler) httperror.LoggerHandler {
	return func(next httperror.LoggerHandler) httperror.LoggerHandler {
		return func(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
			auditor := &Auditor{dataStore: dataStore()}

			payload := replayBody(r, payloadKeys)

			sw := &statusWriter{ResponseWriter: w}
			handlerError := next(sw, r)

			if handlerError != nil {
				payload["error"] = handlerError.Message
			}
			for key, value := range mux.Vars(r) {
				if _, ok := payload[key]; !ok && value != "" {
					payload[key] = value
				}
			}

			auditor.Record(r, auditor.endpointIDFromRequest(r), action, payload, handlerError == nil)

			return handlerError
		}
	}
}

// Record writes a pre-built log entry. Used by the middleware and by the
// Docker proxy transport for streaming actions (image pull).
func (auditor *Auditor) Record(r *http.Request, endpointID portainer.EndpointID, action string, payload map[string]any, success bool) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["success"] = success

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		encodedPayload = []byte("{}")
	}

	entry := &portainer.UserActivityLog{
		Timestamp: time.Now().Unix(),
		Context:   auditor.contextName(endpointID),
		Action:    action,
		Payload:   encodedPayload,
	}

	if tokenData, err := security.RetrieveTokenData(r); err == nil && tokenData != nil {
		entry.Username = tokenData.Username
		entry.APIKeyID = int(tokenData.APIKeyID)
		if tokenData.APIKeyID != 0 {
			if apiKey, err := auditor.dataStore.APIKeyRepository().GetAPIKey(tokenData.APIKeyID); err == nil {
				entry.APIKey = apiKey.Description
			}
		}
	} else {
		entry.Username = "webhook"
	}

	if err := auditor.dataStore.UserActivityLog().Log(entry); err != nil {
		log.Error().Err(err).Str("action", action).Msg("Unable to store user activity log entry")
	}
}

// replayBody buffers and restores the request body, returning the selected
// top-level JSON fields of the body.
func replayBody(r *http.Request, payloadKeys []string) map[string]any {
	payload := map[string]any{}

	if r.Body == nil || r.ContentLength == 0 {
		return payload
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return payload
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return payload
	}

	for _, key := range payloadKeys {
		if value, ok := parsed[key]; ok {
			payload[key] = value
		}
	}

	return payload
}

func (auditor *Auditor) endpointIDFromRequest(r *http.Request) portainer.EndpointID {
	if endpointID, err := strconv.Atoi(r.URL.Query().Get("endpointId")); err == nil && endpointID != 0 {
		return portainer.EndpointID(endpointID)
	}

	vars := mux.Vars(r)
	if endpointID, err := strconv.Atoi(vars["id"]); err == nil && endpointID != 0 {
		return portainer.EndpointID(endpointID)
	}

	if webhookID := vars["webhookID"]; webhookID != "" {
		stack, err := auditor.dataStore.Stack().StackByWebhookID(webhookID)
		if err == nil {
			return stack.EndpointID
		}
	}

	return 0
}

func (auditor *Auditor) contextName(endpointID portainer.EndpointID) string {
	if endpointID == 0 {
		return "portainer"
	}

	endpoint, err := auditor.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return fmt.Sprintf("environment %d", endpointID)
	}

	return endpoint.Name
}
