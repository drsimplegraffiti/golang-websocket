package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/models"
	"golangchatapp/internal/utils"
)

func handleGetPrivate(w http.ResponseWriter, r *http.Request) {
	privateIdStr := r.PathValue("private_id")
	privateId, err := strconv.ParseInt(privateIdStr, 10, 64)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid private_id", nil)
		return
	}

	private, err := models.GetPrivateById(privateId)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "failed to fetch private by id", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "Private fetched", private)
}

func handleCreatePrivate(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.CtxUserID).(int64)

	var req struct {
		ReceiverId int64 `json:"receiver_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ReceiverId == 0 {
		utils.JSON(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	private, err := models.GetPrivateByUsers(userId, req.ReceiverId)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	if private == nil {
		private, err = models.CreatePrivate(userId, req.ReceiverId)
		if err != nil {
			utils.JSON(w, http.StatusInternalServerError, false, err.Error(), nil)
			return
		}
		utils.JSON(w, http.StatusCreated, true, "private chat created", private)
		return
	}

	utils.JSON(w, http.StatusCreated, true, "private chat already exists", private)
}

func handleGetConversations(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.CtxUserID).(int64)

	privates, err := models.GetPrivatesForUser(userId)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	utils.JSON(w, http.StatusCreated, true, "privated created", privates)
}

func handleGetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	privateIdStr := r.PathValue("private_id")
	privateId, err := strconv.ParseInt(privateIdStr, 10, 64)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid private_id", nil)
		return
	}

	page := 1
	limit := 20

	p := r.URL.Query().Get("page")
	if p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v <= 0 {
			utils.JSON(w, http.StatusBadRequest, false, "invalid page query", nil)
			return
		}
		page = v
	}

	l := r.URL.Query().Get("limit")
	if l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v <= 0 {
			utils.JSON(w, http.StatusBadRequest, false, "invalid limit query", nil)
			return
		}
		limit = v
	}

	messages, err := models.GetMessagesByPrivateID(privateId, page, limit+1)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "failed to fetch private messages", nil)
		return
	}

	hasNextPage := false
	if len(messages) > limit {
		hasNextPage = true
		messages = messages[:limit]
	}

	utils.JSON(w, http.StatusOK, true, "private messages fetched", map[string]any{
		"messages":      messages,
		"page":          page,
		"limit":         limit,
		"has_next_page": hasNextPage,
	})
}
