package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id UserInspectCurrentAPIKey
// @summary Inspect the API key used for the current request
// @description Retrieve the API key that authenticated the current request.
// @description This endpoint only returns a key when the request is authenticated with X-API-Key.
// @description The response includes effectiveAccessPreset, which includes any active temporary elevation.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @produce json
// @success 200 {object} apiKeyResponse "Success"
// @failure 400 "Current request was not authenticated with an API key"
// @failure 403 "Permission denied"
// @failure 404 "API key not found"
// @failure 500 "Server error"
// @router /users/me/current-api-key [get]
func (handler *Handler) userInspectCurrentAPIKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.APIKeyID == 0 {
		return httperror.BadRequest("Current request was not authenticated with an API key", errors.New("missing API key context"))
	}

	apiKey, err := handler.apiKeyService.GetAPIKey(tokenData.APIKeyID)
	if err != nil {
		return httperror.NotFound("Unable to find the current API key", err)
	}

	if apiKey.UserID != tokenData.ID {
		return httperror.Forbidden("Permission denied to inspect current API key", errors.New("API key does not belong to the authenticated user"))
	}

	return response.JSON(w, buildAccessTokenResponse(apiKey, time.Now().UTC().Unix()))
}
