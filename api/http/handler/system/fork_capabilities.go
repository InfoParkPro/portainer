package system

import (
	"net/http"

	"github.com/portainer/portainer/api/internal/forkdocs"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

func (handler *Handler) forkCapabilities(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	return response.JSON(w, forkdocs.Capabilities())
}
