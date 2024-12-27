package api

import (
	"net/http"

	"github.com/erikgmatos/gobid/internal/jsonutils"
	"github.com/gorilla/csrf"
)

func (api *Api) HandleGetCsrfToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{"csrf_token": token})
}

func (api *Api) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Sessions.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]string{"error": "must be logged in"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
