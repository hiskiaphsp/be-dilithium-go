package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vercel/go-bridge/go/bridge"

	"be-dilithium/config"
	"be-dilithium/controllers"
	controller "be-dilithium/controllers"
	"be-dilithium/repositories"
	"be-dilithium/services"
	service "be-dilithium/services"
)

type handler struct{}

// ServeHTTP implements the http.Handler interface
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure MongoDB connection is available
	if config.MongoDB == nil {
		log.Fatal("MongoDB connection is not available")
	}

	// Create database and collection instances
	db := config.MongoDB.Database(config.DBName)
	documentRepo := repositories.NewDocumentRepository(db, "documents")
	documentService := services.NewDocumentService(documentRepo)
	documentController := controllers.NewDocumentController(documentService, os.Getenv("PUBLIC_STORAGE"))
	dilithiumController := controller.NewDilithiumController(service.NewDilithiumService)

	// Initialize Gin router
	router := gin.New()

	router.Static("/public/storage", "public/storage")

	// Create API v1 group
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/generate-keypair", dilithiumController.GenerateKeyPair)
		apiV1.POST("/sign-message", dilithiumController.SignMessage)
		apiV1.POST("/sign-message-url", dilithiumController.SignMessageUrl)
		apiV1.POST("/verify-signature", dilithiumController.VerifySignature)
		apiV1.POST("/verify-signature-url", dilithiumController.VerifySignatureURL)
		apiV1.POST("/analyze", dilithiumController.AnalyzeExecutionTimeAndSizes)
		apiV1.POST("/analyze-url", dilithiumController.AnalyzeExecutionTimeAndSizesUrl)

		// Document routes
		apiV1.POST("/documents", documentController.CreateDocument)
		apiV1.GET("/documents/:id", documentController.GetDocumentByID)
		apiV1.GET("/documents", documentController.GetAllDocuments)
		apiV1.PUT("/documents", documentController.UpdateDocument)
		apiV1.DELETE("/documents/:id", documentController.DeleteDocument)
	}

	router.ServeHTTP(w, r)
}

func main() {
	bridge.Start(&handler{})
}
