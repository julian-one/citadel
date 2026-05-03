package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/broker"
)

func GetTradingAccount(l *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := b.GetAccount()
		if err != nil {
			l.Error("failed to get trading account", "error", err)
			http.Error(w, "Failed to retrieve trading account", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	}
}
