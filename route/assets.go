package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/broker"
)

func SearchAssets(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		assets, err := b.SearchAssets(q)
		if err != nil {
			logger.Error("failed to search assets", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to search assets"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(assets)
	}
}
