package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func calculateAspectRatio(width, height int) string {
	if width == 0 || height == 0 {
		return "other"
	}
	ratio := float64(width) / float64(height)
	const epsilon = 0.01

	switch {
	case math.Abs(ratio-16.0/9.0) < epsilon:
		return "16:9"
	case math.Abs(ratio-9.0/16.0) < epsilon:
		return "9:16"
	default:
		return "other"
	}
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	var ffp FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &ffp); err != nil {
		return "", err
	}
	if len(ffp.Streams) == 0 {
		return "", fmt.Errorf("no streams found")
	}
	return calculateAspectRatio(ffp.Streams[0].Width, ffp.Streams[0].Height), nil

}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := fmt.Sprintf("%s.processing", filePath)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputPath, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	presignedRequest, err := presignClient.PresignGetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		},
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", err
	}

	return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	parts := strings.Split(*video.VideoURL, ",")
	if len(parts) != 2 {
		return database.Video{}, fmt.Errorf("invalid video URL format: expected 'bucket,key', got %s", *video.VideoURL)
	}
	bucket := parts[0]
	key := parts[1]
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 15*time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	video.VideoURL = &presignedURL
	return video, nil
}

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
	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, 500, "Unable to get aspect ratio for media", nil)
		return
	}
	newfilepath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		respondWithError(w, 500, "Unable to preprocessing the media", err)
		return
	}
	processedFile, err := os.Open(newfilepath)
	if err != nil {
		respondWithError(w, 500, "Could not open processed video file", err)
		return
	}
	defer processedFile.Close()
	defer os.Remove(newfilepath)
	// tmpFile.Seek(0, io.SeekStart)

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

	prefix := "other"
	switch aspectRatio {
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	}

	key := make([]byte, 32)
	rand.Read(key)
	assetPath := fmt.Sprintf("%s/%s.%s", prefix, base64.RawURLEncoding.EncodeToString(key), ext)

	cfg.s3Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket:      aws.String(cfg.s3Bucket),
			Key:         aws.String(assetPath),
			Body:        processedFile,
			ContentType: aws.String(mediaType),
		},
	)
	videoURL := fmt.Sprintf("%s,%s", cfg.s3Bucket, assetPath)
	video.VideoURL = &videoURL
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "Unable to update video", err)
		return
	}
	signedVideo, err := cfg.dbVideoToSignedVideo(video)
	if err != nil {
		respondWithError(w, 500, "Failed to generate presigned URL", err)
		return
	}
	respondWithJSON(w, http.StatusOK, signedVideo)
}
