package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golangchatapp/internal/events"
	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/models"
	"golangchatapp/internal/utils"
)

// func handleEmailRegister(w http.ResponseWriter, r *http.Request) { // *http is
func handleEmailRegister(w http.ResponseWriter, r *http.Request, eventBus *events.EventBus) {
	// used to define a pointer to an http.Request struct, which represents the
	// incoming HTTP request. The http.ResponseWriter is used to construct the HTTP
	// response.
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// json.NewDecoder means that we are creating a new JSON decoder that reads
	// from the request body (r.Body). The Decode method is then called on this
	// decoder, which attempts to parse the JSON data from the request body and
	// populate the req struct with the corresponding values. If the JSON is
	// malformed or doesn't match the expected structure, an error will be
	// returned.

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if req.Email == "" || req.Name == "" || req.Password == "" {
		utils.JSON(w, http.StatusBadRequest, false, "invalid credentials", nil)
		return
	}

	existingUser, _ := models.GetUserByEmail(req.Email)
	if existingUser != nil {
		utils.JSON(w, http.StatusConflict, false, "user already exists", nil)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "signup failed please try again later", nil)
		return
	}

	user, err := models.CreateUserByEmail(req.Name, req.Email, hashedPassword)
	if err != nil {
		fmt.Println(err)
		utils.JSON(w, http.StatusInternalServerError, false, "internal server error", nil)
		return
	}

	// ─── Event-Driven Email Trigger ─────────────────────────────
	// Fire-and-forget: HTTP response isn't blocked by email sending
	go func() {
		// Generate verification URL (you'd have a proper token generation)
		verifyURL := fmt.Sprintf("http://localhost:8080/verify?token=%s&user=%d",
			"temp-token", user.ID)

		eventPayload := struct {
			UserID    int64  `json:"user_id"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			VerifyURL string `json:"verify_url"`
		}{
			UserID:    user.ID,
			Email:     user.Email,
			Name:      user.Name,
			VerifyURL: verifyURL,
		}

		event, err := events.NewEvent(events.EventUserRegistered, eventPayload)
		if err != nil {
			// Log but don't fail the request
			fmt.Printf("Failed to create event: %v\n", err)
			return
		}

		// Use background context with timeout for publishing
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish event: %v\n", err)
		}
	}()

	utils.JSON(w, http.StatusCreated, true, "success", user)
}

func handleEmailLogin(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(strings.TrimSpace(r.Header.Get(middlewares.CtxPlatform)))
	if platform != middlewares.PlatformMobile && platform !=
		middlewares.PlatformWeb {
		utils.JSON(w, http.StatusBadRequest, false, "invalid platform", nil)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.JSON(w, http.StatusBadRequest, false, "invalid credentials", nil)
		return
	}

	existingUser, err := models.GetUserByEmail(req.Email)
	if err != nil || existingUser == nil {
		utils.JSON(w, http.StatusConflict, false, "user already exists", nil)
		return
	}

	err = utils.CheckHashPassword(existingUser.Password, req.Password)
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, false, "invalid credentials", nil)
		return
	}

	accessToken, err := utils.GenerateJWT(existingUser.ID, existingUser.Name, platform)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E0 - login failed", nil)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E1 - login failed", nil)
		return
	}

	err = models.UpdateUserUserRefreshToken(existingUser.ID, platform, refreshToken)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E3 - login failed", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "login successful", map[string]any{
		"user":         existingUser,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middlewares.CtxUserID).(int64)
	if !ok {
		utils.JSON(w, http.StatusUnauthorized, false, "Unauthorized", nil)
		return
	}

	platform, ok := r.Context().Value(middlewares.CtxPlatform).(string)

	if !ok {
		utils.JSON(w, http.StatusUnauthorized, false, "L0 - Unauthorized", nil)
		return
	}

	err := models.DeleteUserRefreshToken(userId, platform)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "something went wrong", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "logged out", nil)
}

func handleRefreshSession(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(strings.TrimSpace(r.Header.Get(middlewares.CtxPlatform)))
	if platform != middlewares.PlatformMobile && platform !=
		middlewares.PlatformWeb {
		utils.JSON(w, http.StatusBadRequest, false, "invalid platform", nil)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if req.RefreshToken == "" {
		utils.JSON(w, http.StatusBadRequest, false, "invalid payload", nil)
		return
	}

	existingUser, err := models.GetUserByRefreshToken(req.RefreshToken, platform)
	if err != nil || existingUser == nil {
		utils.JSON(w, http.StatusConflict, false, "invalid credentials", nil)
		return
	}

	accessToken, err := utils.GenerateJWT(existingUser.ID, existingUser.Name, platform)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E0 - session failed", nil)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E1 - session failed", nil)
		return
	}

	err = models.UpdateUserUserRefreshToken(existingUser.ID, platform, refreshToken)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "E3 - session failed", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "login successful", map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(strings.TrimSpace(r.Header.Get(middlewares.CtxPlatform)))
	if platform != middlewares.PlatformMobile && platform !=
		middlewares.PlatformWeb {
		utils.JSON(w, http.StatusBadRequest, false, "invalid platform", nil)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if req.RefreshToken == "" {
		utils.JSON(w, http.StatusBadRequest, false, "invalid payload", nil)
		return
	}

	existingUser, err := models.GetUserByRefreshToken(req.RefreshToken, platform)
	if err != nil || existingUser == nil {
		utils.JSON(w, http.StatusConflict, false, "invalid credentials", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "user profile", map[string]any{
		"user": existingUser,
	})
}
