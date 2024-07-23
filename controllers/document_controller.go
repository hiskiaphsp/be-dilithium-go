package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"be-dilithium/models"
	"be-dilithium/services"
	"be-dilithium/utils"

	"github.com/gin-gonic/gin"
)

type DocumentController struct {
	Service       *services.DocumentService
	PublicStorage string
}

func NewDocumentController(service *services.DocumentService, publicStorage string) *DocumentController {
	return &DocumentController{Service: service, PublicStorage: publicStorage}
}

func (c *DocumentController) CreateDocument(ctx *gin.Context) {
	file, _ := ctx.FormFile("file")
	filename := file.Filename

	timestamp := time.Now().UnixNano()
	folderName := strconv.FormatInt(timestamp, 10)
	uploadDir := filepath.Join("public/storage", folderName)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create directory", err)
		return
	}

	if err := ctx.SaveUploadedFile(file, filepath.Join(uploadDir, filename)); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload file", err)
		return
	}

	document := &models.Document{
		Filename: filename,
		Path:     filepath.Join(uploadDir, filename),
	}

	doc, err := c.Service.Create(ctx, document)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create document", err)
		return
	}

	utils.CreatedResponse(ctx, "Document uploaded successfully", doc)
}

func (c *DocumentController) GetDocumentByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	document, err := c.Service.GetById(ctx, uint(id))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Document not found", err)
		return
	}

	documentURL := c.PublicStorage + document.Path
	fileInfo, err := os.Stat(document.Path)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get file info", err)
		return
	}
	fileSize := fileInfo.Size()

	utils.SuccessResponse(ctx, "Document retrieved successfully", gin.H{
		"document": document,
		"url":      documentURL,
		"size":     fileSize,
	})
}

func (c *DocumentController) GetAllDocuments(ctx *gin.Context) {
	documents, err := c.Service.GetAll(ctx)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch documents", err)
		return
	}

	var documentsWithUrls []gin.H
	for _, doc := range documents {
		documentURL := c.PublicStorage + doc.Path
		fileInfo, err := os.Stat(doc.Path)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get file info", err)
			return
		}
		fileSize := fileInfo.Size()

		docWithUrl := gin.H{
			"document": doc,
			"url":      documentURL,
			"size":     fileSize,
		}
		documentsWithUrls = append(documentsWithUrls, docWithUrl)
	}

	// Ensure that even if no documents are found, we return an empty slice instead of null
	if documentsWithUrls == nil {
		documentsWithUrls = []gin.H{}
	}

	utils.SuccessResponse(ctx, "Documents retrieved successfully", documentsWithUrls)
}

func (c *DocumentController) UpdateDocument(ctx *gin.Context) {
	var document models.Document
	if err := ctx.ShouldBindJSON(&document); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid document data", err)
		return
	}

	doc, err := c.Service.Update(ctx, &document)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update document", err)
		return
	}

	utils.SuccessResponse(ctx, "Document updated successfully", doc)
}

func (c *DocumentController) DeleteDocument(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	document, err := c.Service.GetById(ctx, uint(id))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Document not found", err)
		return
	}

	if err := os.Remove(document.Path); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete file", err)
		return
	}

	folderPath := filepath.Dir(document.Path)
	if err := os.Remove(folderPath); err != nil && !os.IsNotExist(err) && !os.IsExist(err) {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete folder", err)
		return
	}

	success, err := c.Service.Delete(ctx, uint(id))
	if err != nil || !success {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete document", err)
		return
	}

	utils.SuccessResponse(ctx, "Document deleted successfully", nil)
}
