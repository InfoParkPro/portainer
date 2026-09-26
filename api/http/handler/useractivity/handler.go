package useractivity

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle user activity log operations.
type Handler struct {
	*mux.Router
	DataStore      dataservices.DataStore
	requestBouncer security.BouncerService
}

// NewHandler creates a handler to manage user activity log operations.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Handle("/useractivity/logs",
		bouncer.AdminAccess(httperror.LoggerHandler(h.activityLogs))).Methods(http.MethodGet)
	h.Handle("/useractivity/logs.csv",
		bouncer.AdminAccess(httperror.LoggerHandler(h.activityLogsExport))).Methods(http.MethodGet)

	return h
}

type activityLogsQuery struct {
	offset   int
	limit    int
	sortBy   string
	sortDesc bool
	keyword  string
	after    int
	before   int
}

// @id UserActivityLogs
// @summary List user activity logs
// @description List user activity logs, newest entries by default.
// @description **Access policy**: administrator
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param offset query int false "Logs to skip"
// @param limit query int false "Maximum number of logs to return"
// @param sortBy query string false "Sort logs by Context, Action, Timestamp or Username"
// @param sortDesc query boolean false "Sort order (true for descending)"
// @param keyword query string false "Keyword filter"
// @param after query int false "Start of the time range (unix timestamp in seconds)"
// @param before query int false "End of the time range (unix timestamp in seconds)"
// @success 200 {object} userActivityLogsResponse "Success"
// @failure 500 "Server error"
// @router /useractivity/logs [get]
func (handler *Handler) activityLogs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	query, err := parseQuery(r)
	if err != nil {
		return httperror.BadRequest("Invalid query parameters", err)
	}

	logs, err := handler.DataStore.UserActivityLog().Logs()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user activity logs", err)
	}

	logs = filterLogs(logs, query)
	sortLogs(logs, query)

	responseLogs := logs
	if query.offset < len(responseLogs) {
		responseLogs = responseLogs[query.offset:]
	}
	if query.limit > 0 && query.limit < len(responseLogs) {
		responseLogs = responseLogs[:query.limit]
	}

	return response.JSON(w, userActivityLogsResponse{
		TotalCount: len(logs),
		Logs:       decorateLogs(responseLogs),
	})
}

// @id UserActivityLogsExport
// @summary Export user activity logs as CSV
// @description Export all user activity logs matching the query as a CSV file.
// @description **Access policy**: administrator
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce text/csv
// @param sortBy query string false "Sort logs by Context, Action, Timestamp or Username"
// @param sortDesc query boolean false "Sort order (true for descending)"
// @param keyword query string false "Keyword filter"
// @param after query int false "Start of the time range (unix timestamp in seconds)"
// @param before query int false "End of the time range (unix timestamp in seconds)"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /useractivity/logs.csv [get]
func (handler *Handler) activityLogsExport(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	query, err := parseQuery(r)
	if err != nil {
		return httperror.BadRequest("Invalid query parameters", err)
	}

	logs, err := handler.DataStore.UserActivityLog().Logs()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user activity logs", err)
	}

	logs = filterLogs(logs, query)
	sortLogs(logs, query)

	buf := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buf)

	if err := csvWriter.Write([]string{"date", "username", "environment", "action", "api-key", "success", "payload"}); err != nil {
		return httperror.InternalServerError("Unable to export user activity logs", err)
	}

	for _, log := range logs {
		record := []string{
			time.Unix(log.Timestamp, 0).UTC().Format(time.RFC3339),
			log.Username,
			log.Context,
			log.Action,
			log.APIKey,
			strconv.FormatBool(log.Payload != nil && payloadSucceeded(log.Payload)),
			string(log.Payload),
		}
		if err := csvWriter.Write(record); err != nil {
			return httperror.InternalServerError("Unable to export user activity logs", err)
		}
	}
	csvWriter.Flush()

	filename := fmt.Sprintf("portainer-user-activity-logs-%s.csv", time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "text/csv")

	if _, err := w.Write(buf.Bytes()); err != nil {
		return httperror.InternalServerError("Unable to export user activity logs", err)
	}

	return nil
}

type activityLogResponse struct {
	ID        int    `json:"id" example:"1"`
	Timestamp int64  `json:"timestamp" example:"1710000000"`
	Username  string `json:"username" example:"admin"`
	Context   string `json:"context" example:"my-environment"`
	Action    string `json:"action" example:"stack_update"`
	Payload   string `json:"payload"`
	APIKeyID  int    `json:"apiKeyId" example:"1"`
	APIKey    string `json:"apiKey" example:"ci-bot"`
}

type userActivityLogsResponse struct {
	Logs       []activityLogResponse `json:"logs"`
	TotalCount int                   `json:"totalCount" example:"1"`
}

func parseQuery(r *http.Request) (activityLogsQuery, error) {
	query := activityLogsQuery{}

	var err error
	if query.offset, err = request.RetrieveNumericQueryParameter(r, "offset", true); err != nil {
		return query, err
	}
	if query.limit, err = request.RetrieveNumericQueryParameter(r, "limit", true); err != nil {
		return query, err
	}
	if query.after, err = request.RetrieveNumericQueryParameter(r, "after", true); err != nil {
		return query, err
	}
	if query.before, err = request.RetrieveNumericQueryParameter(r, "before", true); err != nil {
		return query, err
	}

	query.sortBy = r.URL.Query().Get("sortBy")
	query.keyword = r.URL.Query().Get("keyword")

	if sortDesc, err := request.RetrieveBooleanQueryParameter(r, "sortDesc", true); err == nil {
		query.sortDesc = sortDesc
	}

	return query, nil
}

func filterLogs(logs []portainer.UserActivityLog, query activityLogsQuery) []portainer.UserActivityLog {
	keyword := strings.ToLower(query.keyword)

	filtered := make([]portainer.UserActivityLog, 0, len(logs))
	for _, log := range logs {
		if query.after != 0 && log.Timestamp < int64(query.after) {
			continue
		}
		if query.before != 0 && log.Timestamp > int64(query.before) {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(
			log.Username+" "+log.Context+" "+log.Action+" "+log.APIKey+" "+string(log.Payload),
		), keyword) {
			continue
		}
		filtered = append(filtered, log)
	}

	return filtered
}

func sortLogs(logs []portainer.UserActivityLog, query activityLogsQuery) {
	sort.SliceStable(logs, func(i, j int) bool {
		a, b := logs[i], logs[j]
		var less bool
		switch query.sortBy {
		case "Username":
			less = a.Username < b.Username
		case "Context":
			less = a.Context < b.Context
		case "Action":
			less = a.Action < b.Action
		default:
			less = a.Timestamp < b.Timestamp
		}
		if query.sortDesc {
			return !less
		}
		return less
	})
}

func decorateLogs(logs []portainer.UserActivityLog) []activityLogResponse {
	response := make([]activityLogResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, activityLogResponse{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			Username:  log.Username,
			Context:   log.Context,
			Action:    log.Action,
			Payload:   base64.StdEncoding.EncodeToString(log.Payload),
			APIKeyID:  log.APIKeyID,
			APIKey:    log.APIKey,
		})
	}
	return response
}

func payloadSucceeded(payload []byte) bool {
	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return false
	}
	return parsed.Success
}
