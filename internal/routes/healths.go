package routes

import (
	"net/http"

	"golangchatapp/internal/utils"
)

func HandleHealthCheckHTTP(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, true, "Api is running!!!", nil)
}
