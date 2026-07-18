package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	// we can use the config package to load the configuration from the
	// environment variables or a .env file
	cfg := config.LoadConfig()

	utils.InitJwt(cfg.JWTKey)

	// db.InitDB(cfg.DBPath, cfg.DBName)
	// defer db.CloseDB() // defer the closing of the database connection until the
	// main function exits

	dbFile := filepath.Join(cfg.DBPath, cfg.DBName)
	log.Println("db file:", dbFile)
	dbConn, err := db.InitDB(cfg.DBPath, cfg.DBName)
	if err != nil {
		log.Fatalf("startup: %v", err)
	}
	defer db.CloseDB(dbConn)

	if err := db.RunMigrations(dbFile, "./migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// logger
	// mux := routes.RegisterRoutes()
	hub := realtime.NewHub()
	mux := routes.RegisterRoutes(hub)

	// observe the chain of middlewares applied to the mux. The order of
	// middleware application is important, as it determines the sequence in
	// which requests are processed. In this case, the LoggingMiddleware is
	// applied first, followed by the CorsMiddleware. This means that every
	// incoming request will first be logged and then have CORS headers applied
	// before reaching the actual route handlers.
	loggerMux := middlewares.LoggingMiddleware(mux)
	corsMux := middlewares.CorsMiddleware(loggerMux)
	// other middlewares to consider: AuthenticationMiddleware,
	// RateLimitingMiddleware, PanicRecoveryMiddleware, etc.

	// mount routes

	server := &http.Server{ // we use & because we want to create a pointer to
		// the http.Server struct, which allows us to modify its fields and call
		// its methods directly.
		Addr: cfg.HTTPServer.Address, // this can be cfg.Address directly,
		// but using cfg.HTTPServer.Address allows for more flexibility in the
		// future if we want to add more server configurations
		Handler: corsMux,

		// other performance tuning options to consider: ReadTimeout,
		// WriteTimeout, IdleTimeout, MaxHeaderBytes, etc.
		MaxHeaderBytes: 1 << 20, // 1 MB
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		ReadTimeout:    15 * time.Second, // this is the maximum duration for
		// reading the entire request, including the body
	}

	shutdownCh := make(chan os.Signal, 1) // make a channel to listen for OS
	// signals, with a buffer size of 1 to ensure that we don't miss any signals
	// if the channel is not read immediately
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server is running on http://%s", cfg.HTTPServer.Address)
		// log.Printf("Server is running at: http://%s", server.Addr)
		// Health Checks
		// log.Printf("Health Check HTTP, GET: http://%s/api/health-check-http", server.Addr)
		// log.Printf("Health Check WS, GET: ws://%s/api/health-check-ws", server.Addr)
		// // Authentications
		// log.Printf("Email register, POST: http://%s/api/auth/register-email", server.Addr)
		// log.Printf("Email login, POST: http://%s/api/auth/login-email", server.Addr)
		// log.Printf("Logout, POST: http://%s/api/auth/logout", server.Addr)
		// log.Printf("Session Refresh, POST: http://%s/api/auth/refresh-session", server.Addr)
		// log.Printf("Current User, POST: http://%s/api/auth/current-user", server.Addr)
		// // Users
		// log.Printf("Get User by ID, GET: http://%s/api/users/{user_id}", server.Addr)
		// // Conversations
		// log.Printf("GET Conversation, GET: http://%s/api/conversations/privates/{private_id}", server.Addr)
		// log.Printf("Join Conversation, POST: http://%s/api/conversations/privates/join", server.Addr)
		// log.Printf("GET All Conversations: http://%s/api/conversations", server.Addr)
		// log.Printf("GET Conversation Messages (paginated): http://%s/api/conversations/privates/{private_id}/messages?page=1&limit=20", server.Addr)
		// WebSocket
		log.Printf("Websocket connection, GET: ws://%s/api/ws", server.Addr)

		err := server.ListenAndServe() // ListenAndServe starts the HTTP server
		// and blocks until the server is stopped or an error occurs. It returns
		// an error if the server fails to start or if it is shut down
		// unexpectedly.
		if err != nil && err != http.ErrServerClosed { // != nil means that an
			// error occurred, and != http.ErrServerClosed means that the
			// error is not due to the server being shut down gracefully. In
			// this case, we log the error and exit the application with a non-zero status code
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	sig := <-shutdownCh // we block the main goroutine until we receive a signal
	// from the shutdownCh
	// i.e when the user presses Ctrl+C or the process receives a termination
	// signal
	log.Printf("Shutdown signal received: %v", sig)

	// context are used to manage timeouts and cancellations for operations that
	// may take a long time to complete. In this case, we create a context with
	// a timeout of 10 seconds, which means that if the server does not shut
	// down gracefully within 10 seconds, the shutdown process will be
	// forcefully terminated.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // the cancel function is called to release resources
	// associated with the context when the shutdown process is complete. This
	// is important to prevent memory leaks and ensure that the application
	// cleans up properly.

	// err := server.Shutdown(ctx) // this method gracefully shuts down the server
	err = server.Shutdown(ctx)

	// without interrupting any active connections. It stops accepting new
	// requests and waits for existing connections to finish within the timeout
	// specified by the context. If the timeout is reached before all connections are closed,
	// the server will forcefully terminate any remaining connections.
	if err != nil {
		log.Printf("server Shutdown failed %v:", err)
	} else {
		log.Println("server Shutdown gracefully")
	}

	hub.Shutdown()

	signal.Stop(shutdownCh) // this stops the signal notification for the
	// shutdownCh channel, preventing further signals from being sent to it. This is important to avoid
	close(shutdownCh)

	log.Println("Application exited cleanly")
}
