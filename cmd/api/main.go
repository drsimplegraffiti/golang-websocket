package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golangchatapp/internal/config"
	"golangchatapp/internal/db"
	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/realtime"
	"golangchatapp/internal/routes"
	"golangchatapp/internal/utils"
)

// Run app with: go run ./cmd/api -config ./config/dev.env
func main() {
	cfg := config.LoadConfig()

	utils.InitJwt(cfg.JWTKey)

	db.InitDB(cfg.DBPath, cfg.DBName)
	defer db.CloseDB()

	// logger
	// mux := routes.RegisterRoutes()
	hub := realtime.NewHub()
	mux := routes.RegisterRoutes(hub)

	loggerMux := middlewares.LoggingMiddleware(mux)
	corsMux := middlewares.CorsMiddleware(loggerMux)

	// mount routes

	server := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: corsMux,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server is running on http://%s", cfg.HTTPServer.Address)
		log.Printf("Server is running at: http://%s", server.Addr)
		// Health Checks
		log.Printf("Health Check HTTP, GET: http://%s/api/health-check-http", server.Addr)
		log.Printf("Health Check WS, GET: ws://%s/api/health-check-ws", server.Addr)
		// Authentications
		log.Printf("Email register, POST: http://%s/api/auth/register-email", server.Addr)
		log.Printf("Email login, POST: http://%s/api/auth/login-email", server.Addr)
		log.Printf("Logout, POST: http://%s/api/auth/logout", server.Addr)
		log.Printf("Session Refresh, POST: http://%s/api/auth/refresh-session", server.Addr)
		log.Printf("Current User, POST: http://%s/api/auth/current-user", server.Addr)
		// Users
		log.Printf("Get User by ID, GET: http://%s/api/users/{user_id}", server.Addr)
		// Conversations
		log.Printf("GET Conversation, GET: http://%s/api/conversations/privates/{private_id}", server.Addr)
		log.Printf("Join Conversation, POST: http://%s/api/conversations/privates/join", server.Addr)
		log.Printf("GET All Conversations: http://%s/api/conversations", server.Addr)
		log.Printf("GET Conversation Messages (paginated): http://%s/api/conversations/privates/{private_id}/messages?page=1&limit=20", server.Addr)
		// WebSocket
		log.Printf("Websocket connection, GET: ws://%s/api/ws", server.Addr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	sig := <-shutdownCh
	log.Printf("Shutdown signal received: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("server Shutdown failed %v:", err)
	} else {
		log.Println("server Shutdown gracefully")
	}

	hub.Shutdown()

	signal.Stop(shutdownCh)
	close(shutdownCh)

	log.Println("Application exited cleanly")
}
