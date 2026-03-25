package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// initRouter initializes the HTTP router for the plugin.
func (p *Plugin) initRouter() *mux.Router {
	router := mux.NewRouter()

	// Middleware to require that the user is logged in
	router.Use(p.MattermostAuthorizationRequired)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/status", p.getStatus).Methods(http.MethodGet)

	return router
}

// ServeHTTP exposes simple plugin HTTP endpoints.
// The root URL is currently <siteUrl>/plugins/com.mattermost.echosummary/api/v1/.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) MattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) getStatus(w http.ResponseWriter, r *http.Request) {
	cfg := p.getConfiguration().normalized()

	response := map[string]any{
		"status":           "ok",
		"plugin_id":        pluginID,
		"configured":       cfg.isConfigured(),
		"bot_user_id":      p.botUserID,
		"timezone":         cfg.NotificationTimezone,
		"default_times":    cfg.DefaultTimeSlots,
		"target_usernames": cfg.TargetUsernames,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to write response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
