package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/hiskiaphsp/be-dilithium-go/config"
	"github.com/hiskiaphsp/be-dilithium-go/controllers"
	controller "github.com/hiskiaphsp/be-dilithium-go/controllers"
	"github.com/hiskiaphsp/be-dilithium-go/repositories"
	"github.com/hiskiaphsp/be-dilithium-go/services"
	service "github.com/hiskiaphsp/be-dilithium-go/services"
)

var router *gin.Engine

func init() {
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
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()

	router.Static("public/storage", "public/storage")

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
}

// Handler is the exported function Vercel looks for
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
