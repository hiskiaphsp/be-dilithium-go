package routes

import (
	"be-dilithium/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterDocumentRoutes(router *gin.RouterGroup, documentController *controllers.DocumentController) {
	router.POST("/documents", documentController.CreateDocument)
	router.GET("/documents/:id", documentController.GetDocumentByID)
	router.GET("/documents", documentController.GetAllDocuments)
	router.PUT("/documents/:id", documentController.UpdateDocument)
	router.DELETE("/documents/:id", documentController.DeleteDocument)
}
