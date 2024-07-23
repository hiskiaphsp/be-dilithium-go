package controllers

import (
	"be-dilithium/models"
	"be-dilithium/services"
	"be-dilithium/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type KeyPairController struct {
	keyPairService *services.KeyPairService
}

// NewKeyPairController creates a new instance of KeyPairController
func NewKeyPairController(keyPairService *services.KeyPairService) *KeyPairController {
	return &KeyPairController{keyPairService: keyPairService}
}

// Create handles creating a new key pair
func (ctrl *KeyPairController) Create(c *gin.Context) {
	var keyPair models.KeyPair
	if err := c.ShouldBindJSON(&keyPair); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	createdKeyPair, err := ctrl.keyPairService.Create(c.Request.Context(), &keyPair)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create key pair", err)
		return
	}

	utils.SuccessResponse(c, "Key pair created successfully", createdKeyPair)
}

// GetById handles fetching a key pair by ID
func (ctrl *KeyPairController) GetById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	keyPair, err := ctrl.keyPairService.GetById(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Key pair not found", err)
		return
	}

	utils.SuccessResponse(c, "Key pair retrieved successfully", keyPair)
}

// GetAll handles fetching all key pairs or searching by field
func (ctrl *KeyPairController) GetAll(c *gin.Context) {
	fields := c.QueryMap("field")

	var keyPairs []models.KeyPair
	var err error

	if len(fields) > 0 {
		for field, value := range fields {
			keyPairs, err = ctrl.keyPairService.FindByField(c, field, value)
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search key pairs", err)
				return
			}
			break // Only support querying by one field at a time
		}
	} else {
		keyPairs, err = ctrl.keyPairService.GetAll(c.Request.Context())
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch key pairs", err)
			return
		}
	}

	utils.SuccessResponse(c, "Key pairs retrieved successfully", keyPairs)
}

// Update handles updating an existing key pair
func (ctrl *KeyPairController) Update(c *gin.Context) {
	var keyPair models.KeyPair
	if err := c.ShouldBindJSON(&keyPair); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updatedKeyPair, err := ctrl.keyPairService.Update(c.Request.Context(), &keyPair)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update key pair", err)
		return
	}

	utils.SuccessResponse(c, "Key pair updated successfully", updatedKeyPair)
}

// Delete handles deleting a key pair by ID
func (ctrl *KeyPairController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	success, err := ctrl.keyPairService.Delete(c.Request.Context(), uint(id))
	if err != nil || !success {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete key pair", err)
		return
	}

	utils.SuccessResponse(c, "Key pair deleted successfully", nil)
}
