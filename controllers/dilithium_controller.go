package controllers

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	service "be-dilithium/services"

	"github.com/gin-gonic/gin"
)

type DilithiumController struct {
	serviceFactory func(modeName string) (*service.DilithiumService, error)
}

func NewDilithiumController(serviceFactory func(modeName string) (*service.DilithiumService, error)) *DilithiumController {
	return &DilithiumController{serviceFactory: serviceFactory}
}

func (ctrl *DilithiumController) GenerateKeyPair(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}

	// Attempt to bind JSON, but ignore errors
	_ = c.ShouldBindJSON(&req)

	// Use Dilithium2 as default mode if not specified
	if req.Mode == "" {
		req.Mode = "Dilithium2"
	}

	dilithiumService, err := ctrl.serviceFactory(req.Mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	publicKey, privateKey, err := dilithiumService.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add the public key file to the zip
	pubFile, err := zipWriter.Create("publicKey.key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := pubFile.Write(publicKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add the private key file to the zip
	privFile, err := zipWriter.Create("privateKey.key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := privFile.Write(privateKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Close the zip writer to flush the buffer
	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Serve the file
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=keys.zip")
	c.Writer.Header().Set("Content-Type", "application/zip")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(buf.Bytes())
}

func (ctrl *DilithiumController) SignMessage(c *gin.Context) {
	mode := c.PostForm("mode")
	if mode == "" {
		mode = "Dilithium2"
	}

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get private key file
	privateKeyFile, err := c.FormFile("privateKey")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private key file is required"})
		return
	}

	privateKeyFileContent, err := privateKeyFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer privateKeyFileContent.Close()

	privateKeyBytes, err := ioutil.ReadAll(privateKeyFileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get message file
	messageFile, err := c.FormFile("message")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message file is required"})
		return
	}

	messageFileContent, err := messageFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer messageFileContent.Close()

	messageBytes, err := ioutil.ReadAll(messageFileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	signature, err := dilithiumService.SignMessage(privateKeyBytes, messageBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create temporary file for the signature
	signatureFile, err := ioutil.TempFile("", "signature-*.sig")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(signatureFile.Name())

	// Write the signature to the temporary file
	if _, err := signatureFile.Write(signature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	signatureFile.Close()

	// Serve the signature file
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(signatureFile.Name()))
	c.File(signatureFile.Name())
}

func (ctrl *DilithiumController) VerifySignature(c *gin.Context) {
	mode := c.PostForm("mode")
	if mode == "" {
		mode = "Dilithium2"
	}

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get public key file
	publicKeyFile, err := c.FormFile("publicKey")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key file is required"})
		return
	}

	publicKeyFileContent, err := publicKeyFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer publicKeyFileContent.Close()

	publicKeyBytes, err := ioutil.ReadAll(publicKeyFileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get message file
	messageFile, err := c.FormFile("message")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message file is required"})
		return
	}

	messageFileContent, err := messageFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer messageFileContent.Close()

	messageBytes, err := ioutil.ReadAll(messageFileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get signature file
	signatureFile, err := c.FormFile("signature")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature file is required"})
		return
	}

	signatureFileContent, err := signatureFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer signatureFileContent.Close()

	signatureBytes, err := ioutil.ReadAll(signatureFileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	valid, err := dilithiumService.VerifySignature(publicKeyBytes, messageBytes, signatureBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON response with validity
	c.JSON(http.StatusOK, gin.H{"valid": valid})
	print(valid)
}
