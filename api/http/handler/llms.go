package handler

import (
	"fmt"
	"net/http"

	"github.com/portainer/portainer/api/internal/forkdocs"
)

func serveLLMSText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, forkdocs.LLMSText())
}
