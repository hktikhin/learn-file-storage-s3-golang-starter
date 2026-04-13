package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64((1 << 30)))
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
	file, header, err := r.FormFile("video")
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
	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, 500, "Unable to create temp file for media", nil)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	io.Copy(tmpFile, file)
	tmpFile.Seek(0, io.SeekStart)

	ext := "mp4"

	switch mediaType {
	case "video/mp4":
		ext = "mp4"
	case "video/webm":
		ext = "webm"
	case "video/ogg":
		ext = "ogv"
	case "video/quicktime":
		ext = "mov"
	case "video/x-msvideo":
		ext = "avi"
	case "video/x-matroska":
		ext = "mkv"
	case "video/x-flv":
		ext = "flv"
	case "video/mpeg":
		ext = "mpeg"
	default:
		if !strings.HasPrefix(mediaType, "video/") {
			respondWithError(w, http.StatusBadRequest, "Invalid media type: only videos are allowed", nil)
			return
		}
	}
	key := make([]byte, 32)
	rand.Read(key)
	assetPath := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(key), ext)

	cfg.s3Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket:      aws.String(cfg.s3Bucket),
			Key:         aws.String(assetPath),
			Body:        tmpFile,
			ContentType: aws.String(mediaType),
		},
	)
	videoURL := fmt.Sprintf("https://sturdy-palm-tree-9wx4q976p99cwj-4566.app.github.dev/%s/%s", cfg.s3Bucket, assetPath)
	video.VideoURL = &videoURL
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "Unable to update video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)
}
