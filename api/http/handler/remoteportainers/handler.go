package remoteportainers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	remoteclient "github.com/portainer/portainer/api/remoteportainer/client"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
	DataStore dataservices.DataStore
}

type remotePortainerPayload struct {
	Name          string
	URL           string
	APIToken      string
	TLSSkipVerify bool
}

type testConnectionResponse struct {
	Status  string `json:"Status"`
	Version string `json:"Version"`
}

func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore) *Handler {
	h := &Handler{
		Router:    mux.NewRouter(),
		DataStore: dataStore,
	}

	h.Handle("/remote_portainers", bouncer.AdminAccess(httperror.LoggerHandler(h.list))).Methods(http.MethodGet)
	h.Handle("/remote_portainers", bouncer.AdminAccess(httperror.LoggerHandler(h.create))).Methods(http.MethodPost)
	h.Handle("/remote_portainers/{id}", bouncer.AdminAccess(httperror.LoggerHandler(h.inspect))).Methods(http.MethodGet)
	h.Handle("/remote_portainers/{id}", bouncer.AdminAccess(httperror.LoggerHandler(h.update))).Methods(http.MethodPut)
	h.Handle("/remote_portainers/{id}", bouncer.AdminAccess(httperror.LoggerHandler(h.delete))).Methods(http.MethodDelete)
	h.Handle("/remote_portainers/{id}/test", bouncer.AdminAccess(httperror.LoggerHandler(h.testConnection))).Methods(http.MethodPost)
	h.Handle("/remote_portainers/{id}/stacks", bouncer.AdminAccess(httperror.LoggerHandler(h.stackList))).Methods(http.MethodGet)
	h.Handle("/remote_portainers/{id}/stacks/{stackId}", bouncer.AdminAccess(httperror.LoggerHandler(h.stackInspect))).Methods(http.MethodGet)
	h.Handle("/remote_portainers/{id}/stacks/{stackId}/file", bouncer.AdminAccess(httperror.LoggerHandler(h.stackFile))).Methods(http.MethodGet)
	h.Handle("/remote_portainers/{id}/stacks/{stackId}", bouncer.AdminAccess(httperror.LoggerHandler(h.stackUpdate))).Methods(http.MethodPut)

	return h
}

func (payload *remotePortainerPayload) Validate(_ *http.Request) error {
	if strings.TrimSpace(payload.Name) == "" {
		return errors.New("name is required")
	}

	if _, err := normalizedURL(payload.URL); err != nil {
		return err
	}

	return nil
}

func (handler *Handler) list(w http.ResponseWriter, _ *http.Request) *httperror.HandlerError {
	remotePortainers, err := handler.DataStore.RemotePortainer().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve remote Portainers from the database", err)
	}

	for i := range remotePortainers {
		remotePortainers[i].APIToken = ""
	}

	return response.JSON(w, remotePortainers)
}

func (handler *Handler) create(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload remotePortainerPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if strings.TrimSpace(payload.APIToken) == "" {
		return httperror.BadRequest("Invalid request payload", errors.New("API token is required"))
	}

	remoteURL, err := normalizedURL(payload.URL)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	now := time.Now().Unix()
	remotePortainer := &portainer.RemotePortainer{
		Name:          strings.TrimSpace(payload.Name),
		URL:           remoteURL,
		APIToken:      payload.APIToken,
		TLSSkipVerify: payload.TLSSkipVerify,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := handler.DataStore.RemotePortainer().Create(remotePortainer); err != nil {
		return httperror.InternalServerError("Unable to persist remote Portainer inside the database", err)
	}

	return response.JSON(w, sanitized(remotePortainer))
}

func (handler *Handler) inspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	remotePortainer, httpErr := handler.remotePortainer(r)
	if httpErr != nil {
		return httpErr
	}

	return response.JSON(w, sanitized(remotePortainer))
}

func (handler *Handler) update(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	remotePortainer, httpErr := handler.remotePortainer(r)
	if httpErr != nil {
		return httpErr
	}

	var payload remotePortainerPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	remoteURL, err := normalizedURL(payload.URL)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	remotePortainer.Name = strings.TrimSpace(payload.Name)
	remotePortainer.URL = remoteURL
	remotePortainer.TLSSkipVerify = payload.TLSSkipVerify
	remotePortainer.UpdatedAt = time.Now().Unix()
	if payload.APIToken != "" {
		remotePortainer.APIToken = payload.APIToken
	}

	if err := handler.DataStore.RemotePortainer().Update(remotePortainer.ID, remotePortainer); err != nil {
		return httperror.InternalServerError("Unable to persist remote Portainer changes inside the database", err)
	}

	return response.JSON(w, sanitized(remotePortainer))
}

func (handler *Handler) delete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, httpErr := remotePortainerID(r)
	if httpErr != nil {
		return httpErr
	}

	if err := handler.DataStore.RemotePortainer().Delete(id); err != nil {
		return httperror.InternalServerError("Unable to delete remote Portainer from the database", err)
	}

	return response.Empty(w)
}

func (handler *Handler) testConnection(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	client, httpErr := handler.client(r)
	if httpErr != nil {
		return httpErr
	}

	status, err := client.Status(r.Context())
	if err != nil {
		return badGateway("Unable to test remote Portainer connection", err)
	}

	return response.JSON(w, testConnectionResponse{
		Status:  "ok",
		Version: status.Version,
	})
}

func (handler *Handler) stackList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	client, httpErr := handler.client(r)
	if httpErr != nil {
		return httpErr
	}

	stacks, err := client.Stacks(r.Context())
	if err != nil {
		return badGateway("Unable to retrieve remote Portainer stacks", err)
	}

	return response.JSON(w, stacks)
}

func (handler *Handler) stackInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	client, httpErr := handler.client(r)
	if httpErr != nil {
		return httpErr
	}

	stackID, httpErr := remoteStackID(r)
	if httpErr != nil {
		return httpErr
	}

	stack, err := client.Stack(r.Context(), stackID)
	if err != nil {
		return badGateway("Unable to retrieve remote Portainer stack", err)
	}

	return response.JSON(w, stack)
}

func (handler *Handler) stackFile(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	client, httpErr := handler.client(r)
	if httpErr != nil {
		return httpErr
	}

	stackID, httpErr := remoteStackID(r)
	if httpErr != nil {
		return httpErr
	}

	content, err := client.StackFile(r.Context(), stackID)
	if err != nil {
		return badGateway("Unable to retrieve remote Portainer stack file", err)
	}

	return response.JSON(w, map[string]string{"StackFileContent": content})
}

func (handler *Handler) stackUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	client, httpErr := handler.client(r)
	if httpErr != nil {
		return httpErr
	}

	stackID, httpErr := remoteStackID(r)
	if httpErr != nil {
		return httpErr
	}

	var payload updateStackPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	stack, err := client.Stack(r.Context(), stackID)
	if err != nil {
		return badGateway("Unable to retrieve remote Portainer stack", err)
	}

	updated, err := client.UpdateStack(r.Context(), stack, remoteclient.UpdateStackPayload(payload))
	if err != nil {
		return badGateway("Unable to update remote Portainer stack", err)
	}

	return response.JSON(w, updated)
}

type updateStackPayload remoteclient.UpdateStackPayload

func (payload *updateStackPayload) Validate(_ *http.Request) error {
	if strings.TrimSpace(payload.StackFileContent) == "" {
		return errors.New("stack file content is required")
	}

	return nil
}

func (handler *Handler) client(r *http.Request) (*remoteclient.Client, *httperror.HandlerError) {
	remotePortainer, httpErr := handler.remotePortainer(r)
	if httpErr != nil {
		return nil, httpErr
	}

	client, err := remoteclient.New(remotePortainer.URL, remotePortainer.APIToken, remotePortainer.TLSSkipVerify)
	if err != nil {
		return nil, httperror.BadRequest("Invalid remote Portainer configuration", err)
	}

	return client, nil
}

func (handler *Handler) remotePortainer(r *http.Request) (*portainer.RemotePortainer, *httperror.HandlerError) {
	id, httpErr := remotePortainerID(r)
	if httpErr != nil {
		return nil, httpErr
	}

	remotePortainer, err := handler.DataStore.RemotePortainer().Read(id)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find the remote Portainer inside the database", err)
	}
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the remote Portainer from the database", err)
	}

	return remotePortainer, nil
}

func remotePortainerID(r *http.Request) (portainer.RemotePortainerID, *httperror.HandlerError) {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return 0, httperror.BadRequest("Invalid remote Portainer identifier route variable", err)
	}

	return portainer.RemotePortainerID(id), nil
}

func remoteStackID(r *http.Request) (portainer.StackID, *httperror.HandlerError) {
	id, err := request.RetrieveNumericRouteVariableValue(r, "stackId")
	if err != nil {
		return 0, httperror.BadRequest("Invalid remote stack identifier route variable", err)
	}

	return portainer.StackID(id), nil
}

func sanitized(remotePortainer *portainer.RemotePortainer) *portainer.RemotePortainer {
	if remotePortainer == nil {
		return nil
	}

	sanitized := *remotePortainer
	sanitized.APIToken = ""
	return &sanitized
}

func normalizedURL(rawURL string) (string, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	if rawURL == "" {
		return "", errors.New("URL is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("URL must include scheme and host")
	}

	return rawURL, nil
}

func badGateway(message string, err error) *httperror.HandlerError {
	return httperror.NewError(http.StatusBadGateway, message, err)
}
