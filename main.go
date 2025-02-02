package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/LCGant/go-transfer-files/controllers"
	"github.com/LCGant/go-transfer-files/models"
)

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/files?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	fmt.Println("Database connection established successfully!")

	if err := db.AutoMigrate(
		&models.File{},
		&models.FileData{},
		&models.Download{},
		&models.ScheduledEvent{},
	); err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	router := gin.Default()

	router.Static("/public", "./public")

	router.GET("/", func(c *gin.Context) {
		c.File("./public/choice.html")
	})

	router.GET("/Files/upload", func(c *gin.Context) {
		c.File("./public/upload.html")
	})

	router.POST("/Files/upload", controllers.UploadFile(db))
	router.GET("/Files/download", controllers.DownloadFile(db))

	port := "8080"
	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
