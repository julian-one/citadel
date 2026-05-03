package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/broker"
)

func SearchAssets(l *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		assets, err := b.SearchAssets(q)
		if err != nil {
			l.Error("failed to search assets", "error", err)
			http.Error(w, "failed to search assets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(assets)
	}
}
