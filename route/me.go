package route

import (
	"encoding/json"
	"net/http"

	"citadel/internal/middleware"
)

func GetMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(middleware.GetUser(r))
	}
}
