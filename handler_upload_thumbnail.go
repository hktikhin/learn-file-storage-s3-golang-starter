package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to fetch mediaType from file", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", err)
		return
	}
	// image, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Unable to parse image file", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Unable to fetch video from UUID", err)
		return
	}
	if (video == database.Video{}) {
		respondWithError(w, http.StatusNotFound, "Video not found", nil)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You are not the owner of the video", nil)
		return
	}
	// videoThumbnails[videoID] = thumbnail{
	// 	data:      images,
	// 	mediaType: mediaType,
	// }
	// thumbnailURL := fmt.Sprintf("https://sturdy-palm-tree-9wx4q976p99cwj-8091.app.github.dev/api/thumbnails/%s", video.ID.String())
	// thumbnailURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(image))
	ext := "jpg"
	if mediaType == "image/png" {
		ext = "png"
	}
	key := make([]byte, 32)
	rand.Read(key)
	assetPath := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(key), ext)
	filePath := filepath.Join(cfg.assetsRoot, assetPath)
	dst, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create file on disk", err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to save file contents", err)
		return
	}
	thumbnailURL := fmt.Sprintf("https://sturdy-palm-tree-9wx4q976p99cwj-8091.app.github.dev/assets/%s", assetPath)
	video.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "Unable to update video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)
}
