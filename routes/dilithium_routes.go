// routes/dilithium_routes.go

package routes

import (
	"be-dilithium/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterDilithiumRoutes initializes the routes for DilithiumController
func RegisterDilithiumRoutes(router *gin.RouterGroup, dilithiumController *controllers.DilithiumController) {
	router.POST("/generate-keypair", dilithiumController.GenerateKeyPair)
	router.POST("/generate-keypair-time", dilithiumController.GenerateKeyPairTime)
	router.POST("/sign-message", dilithiumController.SignMessage)
	router.POST("/sign-message-url", dilithiumController.SignMessageUrl)
	router.POST("/verify-signature", dilithiumController.VerifySignature)
	router.POST("/verify-signature-url", dilithiumController.VerifySignatureUrl)
}
