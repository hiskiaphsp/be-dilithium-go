package routes

import (
	"be-dilithium/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterKeyPairRoutes sets up routes for KeyPair operations
func RegisterKeyPairRoutes(router *gin.RouterGroup, keyPairController *controllers.KeyPairController) {
	router.POST("/keypairs", keyPairController.Create)
	router.GET("/keypairs/:id", keyPairController.GetById)
	router.GET("/keypairs", keyPairController.GetAll)
	router.PUT("/keypairs/:id", keyPairController.Update)
	router.DELETE("/keypairs/:id", keyPairController.Delete)
}
