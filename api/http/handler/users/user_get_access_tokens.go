package users

import (
	"net/http"
	"time"

	portainer "github.com/portainer/portainer/api"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type apiKeyResponse struct {
	portainer.APIKey
	EffectiveAccessPreset portainer.APIKeyAccessPreset `json:"effectiveAccessPreset" example:"manage"`
}

// @id UserGetAPIKeys
// @summary Get all API keys for a user
// @description Gets all API keys for a user.
// @description Only the calling user or admin can retrieve api-keys.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {array} apiKeyResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/tokens [get]
func (handler *Handler) userGetAccessTokens(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	user, err := handler.DataStore.User().Read(portainer.UserID(userID))
	if err != nil {
		if handler.DataStore.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}

	if tokenData.ID != portainer.UserID(userID) && (tokenData.Role != portainer.AdministratorRole || user.Role == portainer.AdministratorRole) {
		return httperror.Forbidden("Permission denied to get user access tokens", httperrors.ErrUnauthorized)
	}

	apiKeys, err := handler.apiKeyService.GetAPIKeys(portainer.UserID(userID))
	if err != nil {
		return httperror.InternalServerError("Internal Server Error", err)
	}

	now := time.Now().UTC().Unix()
	tokens := make([]apiKeyResponse, len(apiKeys))
	for idx := range apiKeys {
		tokens[idx] = buildAccessTokenResponse(&apiKeys[idx], now)
	}

	return response.JSON(w, tokens)
}

// hideAPIKeyFields remove the digest from the API key (it is not needed in the response)
func hideAPIKeyFields(apiKey *portainer.APIKey) {
	apiKey.Digest = ""
}

func buildAccessTokenResponse(apiKey *portainer.APIKey, now int64) apiKeyResponse {
	hideAPIKeyFields(apiKey)

	return apiKeyResponse{
		APIKey:                *apiKey,
		EffectiveAccessPreset: apiKey.EffectiveAccessPreset(now),
	}
}
