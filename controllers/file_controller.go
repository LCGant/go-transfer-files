package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/LCGant/go-transfer-files/models"
	"github.com/LCGant/go-transfer-files/services"
)

// Maximum allowed size: 100MB
const MaxFileSize = 100 * 1024 * 1024

// UploadFile handles file uploads without authentication and without splitting/encryption.
func UploadFile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		durationStr := c.PostForm("availabilityDuration")
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
			return
		}

		allowedDurations := []int{1, 10, 30, 60, 120, 180}
		if !services.IsValidDuration(duration, allowedDurations) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duration not allowed"})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}
		defer file.Close()

		if header.Size > MaxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File exceeds the maximum allowed size of 100MB"})
			return
		}

		uploadsDir := "uploads"
		if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
			return
		}

		sanitizedName := services.SanitizeFileName(header.Filename)
		storedFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), sanitizedName)
		filePath := filepath.Join(uploadsDir, storedFilename)

		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
			return
		}
		defer out.Close()

		if _, err = io.Copy(out, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error writing the file"})
			return
		}

		fileHash, err := services.GenerateFileHash(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate file hash"})
			return
		}

		downloadToken := uuid.New().String()

		expiryTime := time.Now().Add(time.Duration(duration) * time.Minute)

		newFile := models.File{
			UploadTime:           time.Now(),
			FileType:             header.Header.Get("Content-Type"),
			FileSize:             header.Size,
			AvailabilityDuration: duration,
			FileHash:             fileHash,
			OriginalFilename:     header.Filename,
			StoredFilename:       storedFilename,
			DownloadToken:        downloadToken,
			ExpiryTime:           expiryTime,
		}

		if err := db.Create(&newFile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "File uploaded successfully",
			"download_link": fmt.Sprintf("/Files/download?token=%s", downloadToken),
		})
	}
}

// DownloadFile handles file downloads.
func DownloadFile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		downloadToken := c.Query("token")
		if downloadToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Download token is required"})
			return
		}

		var fileRecord models.File
		if err := db.Where("download_token = ?", downloadToken).First(&fileRecord).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		if time.Now().After(fileRecord.ExpiryTime) {
			c.JSON(http.StatusGone, gin.H{"error": "File expired"})
			return
		}

		filePath := filepath.Join("uploads", fileRecord.StoredFilename)
		f, err := os.Open(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open the file"})
			return
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the file"})
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileRecord.OriginalFilename))
		c.Data(http.StatusOK, fileRecord.FileType, data)
	}
}
