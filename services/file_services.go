package services

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/LCGant/go-transfer-files/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MaxFileSize      = 100 * 1024 * 1024
	MaxTotalFileSize = 20 * 1024 * 1024 * 1024
)

// masterKey to encrypt the file key (optional)
var masterKey []byte

func init() {
	// Read MASTER_ENCRYPTION_KEY from the environment (optional)
	keyBase64 := os.Getenv("MASTER_ENCRYPTION_KEY")
	if keyBase64 == "" {
		fmt.Println("MASTER_ENCRYPTION_KEY variable not defined. Proceeding without key encryption...")
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil || len(decoded) != 32 {
		fmt.Println("Invalid MASTER_ENCRYPTION_KEY or it does not have 32 bytes. Disabling key encryption.")
		return
	}
	masterKey = decoded
}

// ValidateFileSize checks if the file size is within the limit (100MB)
func ValidateFileSize(header *multipart.FileHeader) error {
	if header.Size > MaxFileSize {
		return errors.New("file exceeds the maximum allowed size of 100MB")
	}
	return nil
}

// SanitizeFileName removes dangerous characters from the file name
func SanitizeFileName(filename string) string {
	return filepath.Base(filename)
}

// GenerateFileHash generates a SHA-256 hash of the file
func GenerateFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// GenerateEncryptionKey creates a 32-byte key (AES-256)
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return key, err
}

// EncryptKeyWithMasterKey encrypts the file key using the masterKey (optional)
func EncryptKeyWithMasterKey(data []byte) ([]byte, error) {
	if len(masterKey) == 0 {
		return data, nil
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptKeyWithMasterKey decrypts the file key using the masterKey
func DecryptKeyWithMasterKey(ciphertext []byte) ([]byte, error) {
	if len(masterKey) == 0 {
		return ciphertext, nil
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}

// EncryptFile performs ZIP + AES-CTR + HMAC
func EncryptFile(filePath string, key []byte) (encrypted, iv, mac []byte, err error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	compressed, err := createZip(filePath, fileBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	iv = make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, err
	}
	encrypted = make([]byte, len(compressed))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(encrypted, compressed)

	macHash := hmac.New(sha256.New, key)
	macHash.Write(encrypted)
	mac = macHash.Sum(nil)

	return encrypted, iv, mac, nil
}

// DecryptFile reverses HMAC + AES-CTR and then extracts the ZIP
func DecryptFile(ciphertext, iv, mac, key []byte) ([]byte, error) {
	macHash := hmac.New(sha256.New, key)
	macHash.Write(ciphertext)
	expected := macHash.Sum(nil)
	if !hmac.Equal(expected, mac) {
		return nil, errors.New("HMAC verification failed")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(decrypted, ciphertext)

	original, err := extractZip(decrypted)
	if err != nil {
		return nil, err
	}
	return original, nil
}

// createZip generates an in-memory ZIP
func createZip(filename string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	fw, err := w.Create(filepath.Base(filename))
	if err != nil {
		w.Close()
		return nil, err
	}
	if _, err := fw.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// extractZip extracts an in-memory ZIP
func extractZip(zipData []byte) ([]byte, error) {
	r := bytes.NewReader(zipData)
	zr, err := zip.NewReader(r, int64(len(zipData)))
	if err != nil {
		return nil, err
	}
	if len(zr.File) == 0 {
		return nil, errors.New("empty zip file")
	}
	rc, err := zr.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, rc); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// GetTotalFileSize sums the size (LENGTH(data)) of all files
func GetTotalFileSize(db *gorm.DB) (int64, error) {
	var totalSize int64
	err := db.Model(&models.FileData{}).
		Select("COALESCE(SUM(LENGTH(data)), 0)").
		Scan(&totalSize).Error
	return totalSize, err
}

// CreateDeletionEvent creates or updates a record in the scheduled_events table
func CreateDeletionEvent(db *gorm.DB, fileID uint, expiryTime time.Time) error {
	evtName := fmt.Sprintf("delete_file_%d", fileID)
	ev := models.ScheduledEvent{
		EventName:     evtName,
		FileID:        int(fileID),
		ScheduledTime: expiryTime,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"scheduled_time"}),
	}).Create(&ev).Error
}

// IsValidDuration checks if the duration is allowed
func IsValidDuration(duration int, allowed []int) bool {
	for _, d := range allowed {
		if d == duration {
			return true
		}
	}
	return false
}
