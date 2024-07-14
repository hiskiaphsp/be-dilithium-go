package controllers

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	service "github.com/hiskiaphsp/be-dilithium-go/services"

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

func (ctrl *DilithiumController) SignMessageUrl(c *gin.Context) {
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

	// Get message URL from request
	messageURL := c.PostForm("messageURL")
	if messageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message URL is required"})
		return
	}

	resp, err := http.Get(messageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message file"})
		return
	}

	// Read message content
	messageBytes, err := ioutil.ReadAll(resp.Body)
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

func (ctrl *DilithiumController) VerifySignatureURL(c *gin.Context) {
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

	// Get message URL from request
	messageURL := c.PostForm("messageURL")
	if messageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message URL is required"})
		return
	}

	resp, err := http.Get(messageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message file"})
		return
	}

	// Read message content
	messageBytes, err := ioutil.ReadAll(resp.Body)
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

	// Perform signature verification
	valid, err := dilithiumService.VerifySignature(publicKeyBytes, messageBytes, signatureBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON response with validity
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (ctrl *DilithiumController) AnalyzeExecutionTimeAndSizes(c *gin.Context) {

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

	// Use Dilithium2 as default mode if not specified
	mode := c.DefaultPostForm("mode", "Dilithium5")

	// Measure execution time for key pair generation
	startKeyGen := time.Now()
	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	publicKey, privateKey, err := dilithiumService.GenerateKeyPair()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	keyGenTime := time.Since(startKeyGen)

	// Measure execution time for signing
	startSign := time.Now()
	signature, err := dilithiumService.SignMessage(privateKey, messageBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	signTime := time.Since(startSign)

	// Measure execution time for signature verification
	startVerify := time.Now()
	valid, err := dilithiumService.VerifySignature(publicKey, messageBytes, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	verifyTime := time.Since(startVerify)

	// Get sizes of keys and signature in bytes
	var privateKeySize, publicKeySize, signatureSize int64

	// Dummy key pair generation to get sizes (actual sizes might differ)
	privateKeySize = int64(len(privateKey))
	publicKeySize = int64(len(publicKey))
	signatureSize = int64(len(signature))
	print(valid)
	// Return analysis results
	c.JSON(http.StatusOK, gin.H{
		"key_generation_time":    keyGenTime.Microseconds(),
		"signing_time":           signTime.Microseconds(),
		"verification_time":      verifyTime.Microseconds(),
		"private_key_size_bytes": privateKeySize,
		"public_key_size_bytes":  publicKeySize,
		"signature_size_bytes":   signatureSize,
		"valid":                  valid,
	})
}

func (ctrl *DilithiumController) AnalyzeExecutionTimeAndSizesUrl(c *gin.Context) {

	// Get message URL from request
	messageURL := c.PostForm("messageURL")
	if messageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message URL is required"})
		return
	}

	resp, err := http.Get(messageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message file"})
		return
	}

	// Read message content
	messageBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use Dilithium2 as default mode if not specified
	mode := c.DefaultPostForm("mode", "Dilithium5")

	// Measure execution time for key pair generation
	startKeyGen := time.Now()
	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	publicKey, privateKey, err := dilithiumService.GenerateKeyPair()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	keyGenTime := time.Since(startKeyGen)

	// Measure execution time for signing
	startSign := time.Now()
	signature, err := dilithiumService.SignMessage(privateKey, messageBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	signTime := time.Since(startSign)

	// Measure execution time for signature verification
	startVerify := time.Now()
	valid, err := dilithiumService.VerifySignature(publicKey, messageBytes, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	verifyTime := time.Since(startVerify)

	// Get sizes of keys and signature in bytes
	var privateKeySize, publicKeySize, signatureSize int64

	// Dummy key pair generation to get sizes (actual sizes might differ)
	privateKeySize = int64(len(privateKey))
	publicKeySize = int64(len(publicKey))
	signatureSize = int64(len(signature))
	print(valid)
	// Return analysis results
	c.JSON(http.StatusOK, gin.H{
		"key_generation_time":    keyGenTime.Microseconds(),
		"signing_time":           signTime.Microseconds(),
		"verification_time":      verifyTime.Microseconds(),
		"private_key_size_bytes": privateKeySize,
		"public_key_size_bytes":  publicKeySize,
		"signature_size_bytes":   signatureSize,
		"valid":                  valid,
	})
}
