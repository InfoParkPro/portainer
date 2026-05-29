package users

import (
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type userAccessTokenUpdatePayload struct {
	AccessPreset portainer.APIKeyAccessPreset `json:"accessPreset"`
}

func (payload *userAccessTokenUpdatePayload) Validate(r *http.Request) error {
	switch payload.AccessPreset {
	case portainer.APIKeyAccessPresetDisabled,
		portainer.APIKeyAccessPresetReadOnly,
		portainer.APIKeyAccessPresetPower,
		portainer.APIKeyAccessPresetManage:
		return nil
	default:
		return errors.New("invalid access preset")
	}
}

// @id UserUpdateAPIKey
// @summary Update an api-key associated to a user
// @description Update an api-key associated to a user.
// @description Only the calling user or admin can update api-key.
// @description **Access policy**: authenticated
// @tags users
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "User identifier"
// @param keyID path int true "Api Key identifier"
// @param body body userAccessTokenUpdatePayload true "details"
// @success 200 {object} portainer.APIKey "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /users/{id}/tokens/{keyID} [put]
func (handler *Handler) userUpdateAccessToken(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	apiKeyID, err := request.RetrieveNumericRouteVariableValue(r, "keyID")
	if err != nil {
		return httperror.BadRequest("Invalid api-key identifier route variable", err)
	}

	var payload userAccessTokenUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user authentication token", err)
	}
	if tokenData.Role != portainer.AdministratorRole && tokenData.ID != portainer.UserID(userID) {
		return httperror.Forbidden("Permission denied to update user access token", httperrors.ErrUnauthorized)
	}

	_, err = handler.DataStore.User().Read(portainer.UserID(userID))
	if err != nil {
		if handler.DataStore.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find a user with the specified identifier inside the database", err)
		}
		return httperror.InternalServerError("Unable to find a user with the specified identifier inside the database", err)
	}

	apiKey, err := handler.apiKeyService.GetAPIKey(portainer.APIKeyID(apiKeyID))
	if err != nil {
		return httperror.InternalServerError("API Key not found", err)
	}
	if apiKey.UserID != portainer.UserID(userID) {
		return httperror.Forbidden("Permission denied to update api-key", httperrors.ErrUnauthorized)
	}

	apiKey.AccessPreset = payload.AccessPreset

	if err := handler.apiKeyService.UpdateAPIKey(apiKey); err != nil {
		return httperror.InternalServerError("Unable to update the api-key", err)
	}

	hideAPIKeyFields(apiKey)

	return response.JSON(w, apiKey)
}
