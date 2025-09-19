package httpx

import (
	"net/http"
)

// Ping is a lightweight JSON endpoint; kept in its own file for clarity.
func Ping(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "pong"})
}
