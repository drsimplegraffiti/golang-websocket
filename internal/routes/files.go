package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/utils"
)

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	senderId, ok := r.Context().Value(middlewares.CtxUserID).(int64)
	if !ok {
		utils.JSON(w, http.StatusUnauthorized, false, "Unauthorized", nil)
		return
	}

	privateIdStr := r.PathValue("private_id")
	privateId, err := strconv.ParseInt(privateIdStr, 10, 64)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, false, "invalid private_id", nil)
		return
	}

	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}
	defer file.Close()

	dirPath := filepath.Join("files", "chats", fmt.Sprintf("%d", privateId), fmt.Sprintf("%d", senderId))
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}

	filePath := filepath.Join(dirPath, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file) // the _ will give base64 you can send to provider like cloudinary
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "file upload failed", nil)
		return
	}

	fileUrl := fmt.Sprintf("/files/chats/%d/%d/%s", privateId, senderId, header.Filename)
	utils.JSON(w, http.StatusOK, true, "file saved", fileUrl)
}

func handleGetFile() http.Handler {
	fs := http.FileServer(http.Dir("./files"))
	return http.StripPrefix("/api/files", fs)
}
