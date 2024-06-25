package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"log"
	"os"

	"github.com/gin-gonic/gin"

	"be-dilithium/models"
	"be-dilithium/services"
)

type DocumentController struct {
	Service       *services.DocumentService
	PublicStorage string // Path untuk menyimpan file publik
}

func NewDocumentController(service *services.DocumentService, publicStorage string) *DocumentController {
	return &DocumentController{Service: service, PublicStorage: publicStorage}
}

// CreateDocument menangani request untuk membuat dokumen baru
func (c *DocumentController) CreateDocument(ctx *gin.Context) {
	file, _ := ctx.FormFile("file")
	filename := file.Filename

	// Generate unique folder name based on timestamp
	timestamp := time.Now().UnixNano()
	folderName := strconv.FormatInt(timestamp, 10)

	// Path to save file
	uploadDir := filepath.Join("public/storage", folderName)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatal(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Save file to generated folder
	if err := ctx.SaveUploadedFile(file, filepath.Join(uploadDir, filename)); err != nil {
		log.Fatal(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	document := &models.Document{
		Filename: filename,
		Path:     filepath.Join(uploadDir, filename),
	}

	_, err := c.Service.Create(ctx, document)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Document uploaded successfully"})
}

// GetDocumentByID menangani request untuk mengambil dokumen berdasarkan ID
func (c *DocumentController) GetDocumentByID(ctx *gin.Context) {
	id := ctx.Param("id")

	document, err := c.Service.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Menghasilkan URL langsung dari path file
	documentURL := c.PublicStorage + document.Path

	// Mendapatkan ukuran file
	fileInfo, err := os.Stat(document.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}
	fileSize := fileInfo.Size()

	ctx.JSON(http.StatusOK, gin.H{
		"document": document,
		"url":      documentURL,
		"size":     fileSize,
	})
}

// GetAllDocuments menangani request untuk mengambil semua dokumen
func (c *DocumentController) GetAllDocuments(ctx *gin.Context) {
	documents, err := c.Service.GetAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}

	// Menghasilkan URL langsung dari path file untuk setiap dokumen
	var documentsWithUrls []gin.H
	for _, doc := range documents {
		documentURL := c.PublicStorage + doc.Path

		// Mendapatkan ukuran file
		fileInfo, err := os.Stat(doc.Path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
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

	ctx.JSON(http.StatusOK, documentsWithUrls)
}

// UpdateDocument menangani request untuk memperbarui dokumen
func (c *DocumentController) UpdateDocument(ctx *gin.Context) {
	var document models.Document
	if err := ctx.ShouldBindJSON(&document); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document data"})
		return
	}

	result, err := c.Service.Update(ctx, &document)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// DeleteDocument menangani request untuk menghapus dokumen berdasarkan ID
func (c *DocumentController) DeleteDocument(ctx *gin.Context) {
	id := ctx.Param("id")

	// Cari dokumen berdasarkan ID
	document, err := c.Service.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Hapus file dari sistem file
	if err := os.Remove(document.Path); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Hapus folder jika kosong
	folderPath := filepath.Dir(document.Path)
	if err := os.Remove(folderPath); err != nil {
		// Jika error bukan karena folder tidak kosong, laporkan error
		if !os.IsNotExist(err) && !os.IsExist(err) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
			return
		}
	}

	// Hapus catatan dokumen dari database
	result, err := c.Service.Delete(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
