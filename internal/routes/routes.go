package routes

import (
	"net/http"

	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/realtime"
)

// func RegisterRoutes() *http.ServeMux {
func RegisterRoutes(hub *realtime.Hub) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", HandleHealthCheckHTTP)

	// authentication

	mux.HandleFunc("POST /api/auth/register-email", handleEmailRegister)
	mux.HandleFunc("POST /api/auth/login-email", handleEmailLogin)
	mux.HandleFunc("POST /api/auth/logout", middlewares.Authenticate(
		handleLogout))

	mux.HandleFunc("POST /api/auth/refresh-session", handleRefreshSession)
	mux.HandleFunc("POST /api/auth/current-user", handleCurrentUser)

	mux.HandleFunc("GET /api/users/{id}", middlewares.Authenticate(
		handleGetUserById))

	// Conversations
	mux.HandleFunc("GET /api/conversations/privates/{private_id}", middlewares.Authenticate(handleGetPrivate))
	mux.HandleFunc("POST /api/conversations/privates/create", middlewares.Authenticate(handleCreatePrivate))
	mux.HandleFunc("GET /api/conversations", middlewares.Authenticate(handleGetConversations))
	mux.HandleFunc("GET /api/conversations/privates/{private_id}/messages", middlewares.Authenticate(handleGetPrivateMessages))
	// GET /api/conversations/privates/123?page=2&limit=50

	// files
	mux.HandleFunc("POST /api/files/{private_id}", middlewares.Authenticate(handleFileUpload))
	mux.Handle("GET /api/files/", middlewares.AuthenticateHandler(handleGetFile()))

	// Websockets
	mux.HandleFunc("GET /api/ws", func(w http.ResponseWriter, r *http.Request) { handleWebSocket(hub, w, r) })

	return mux
}
