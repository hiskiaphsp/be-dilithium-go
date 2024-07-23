package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"be-dilithium/config"
	"be-dilithium/controllers"
	"be-dilithium/repositories"
	"be-dilithium/routes"
	"be-dilithium/services"
)

// Handler initializes and runs the Gin server.
func Handler() *gin.Engine {
	// Ensure database connection is available
	if config.DB == nil {
		log.Fatal("Database connection is not available")
	}

	// Create repository and service instances
	documentRepo := repositories.NewDocumentRepository(config.DB)
	documentService := services.NewDocumentService(documentRepo)

	keyPairRepo := repositories.NewKeyPairRepository(config.DB)
	keyPairService := services.NewKeyPairService(keyPairRepo)

	signatureRepo := repositories.NewSignatureRepository(config.DB)
	signatureService := services.NewSignatureService(signatureRepo)
	// signatureController := controllers.NewSignatureController(signatureService)

	//controller init
	documentController := controllers.NewDocumentController(documentService, os.Getenv("PUBLIC_STORAGE"))
	keyPairController := controllers.NewKeyPairController(keyPairService)
	dilithiumController := controllers.NewDilithiumController(services.NewDilithiumService, keyPairService, signatureService)

	// Initialize Gin router
	router := gin.Default()

	router.Static("public/storage", "public/storage")

	// Create API v1 group
	apiV1 := router.Group("/api/v1")
	{
		routes.RegisterDilithiumRoutes(apiV1, dilithiumController)
		routes.RegisterDocumentRoutes(apiV1, documentController)
		routes.RegisterKeyPairRoutes(apiV1, keyPairController)
		// routes.RegisterSignatureRoutes(apiV1, signatureController)
	}

	return router
}

func main() {
	router := Handler()
	// Run server
	port := config.Port
	if port == "" {
		port = "80"
	}
	router.Run(":" + port)
}
