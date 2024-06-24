// main.go

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"be-dilithium/config"
	"be-dilithium/controllers"
	controller "be-dilithium/controllers"
	"be-dilithium/repositories"
	"be-dilithium/services"
	service "be-dilithium/services"
)

func main() {
	// Ensure MongoDB connection is available
	if config.MongoDB == nil {
		log.Fatal("MongoDB connection is not available")
	}

	// Create database and collection instances
	db := config.MongoDB.Database(config.DBName)
	documentRepo := repositories.NewDocumentRepository(db, "documents")
	documentService := services.NewDocumentService(documentRepo)
	documentController := controllers.NewDocumentController(documentService, os.Getenv("PUBLIC_STORAGE"))
	// Initialize the Dilithium controller with a factory function for creating the service
	dilithiumController := controller.NewDilithiumController(service.NewDilithiumService)

	// Initialize Gin router
	router := gin.Default()

	router.Static("public/storage", "public/storage")

	router.POST("/generate-keypair", dilithiumController.GenerateKeyPair)
	router.POST("/sign-message", dilithiumController.SignMessage)
	router.POST("/verify-signature", dilithiumController.VerifySignature)

	// Routes
	router.POST("/documents", documentController.CreateDocument)
	router.GET("/documents/:id", documentController.GetDocumentByID)
	router.GET("/documents", documentController.GetAllDocuments)
	router.PUT("/documents", documentController.UpdateDocument)
	router.DELETE("/documents/:id", documentController.DeleteDocument)

	// Run server
	port := config.Port
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
