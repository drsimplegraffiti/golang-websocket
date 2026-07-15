package routes

import (
	"net/http"
	"strconv"

	"golangchatapp/internal/models"
	"golangchatapp/internal/utils"
)

func handleGetUserById(w http.ResponseWriter, r *http.Request) {
	strId := r.PathValue("id")
	targetId, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "could not parse id", nil)
		return
	}

	existingUser, err := models.GetUserById(targetId)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, false, "user not found", nil)
		return
	}

	utils.JSON(w, http.StatusOK, true, "user found", map[string]any{"user": existingUser})
}
