package controllers

import (
	"archive/zip"
	"be-dilithium/utils"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"be-dilithium/models"
	service "be-dilithium/services"

	"github.com/gin-gonic/gin"
)

type DilithiumController struct {
	serviceFactory   func(modeName string) (*service.DilithiumService, error)
	keyPairService   *service.KeyPairService
	signatureService *service.SignatureService
}

func NewDilithiumController(serviceFactory func(modeName string) (*service.DilithiumService, error), keyPairService *service.KeyPairService, signatureService *service.SignatureService) *DilithiumController {
	return &DilithiumController{
		serviceFactory:   serviceFactory,
		keyPairService:   keyPairService,
		signatureService: signatureService,
	}
}

func (ctrl *DilithiumController) GenerateKeyPairTime(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}

	_ = c.ShouldBindJSON(&req)

	if req.Mode == "" {
		req.Mode = "Dilithium2"
	}

	startTime := time.Now()

	dilithiumService, err := ctrl.serviceFactory(req.Mode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Capture memory usage before key pair generation
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	startGenerate := time.Now()
	publicKey, privateKey, err := dilithiumService.GenerateKeyPair()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	endGenerate := time.Since(startGenerate)

	// Capture memory usage after key pair generation
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage during key pair generation
	memUsage := memStatsAfter.Alloc - memStatsBefore.Alloc

	// Calculate communication size
	communicationSize := len(publicKey) + len(privateKey)

	getKeyHash := func(key []byte) string {
		hash := sha256.Sum256(key)
		return hex.EncodeToString(hash[:])
	}

	pubKeyHashStr := getKeyHash(publicKey)
	privKeyHashStr := getKeyHash(privateKey)

	keyPair := &models.KeyPair{
		PrivateKey: privKeyHashStr,
		PublicKey:  pubKeyHashStr,
		Variant:    req.Mode,
	}

	_, err = ctrl.keyPairService.Create(c.Request.Context(), keyPair)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	pubKeyFileName := "publicKey_" + req.Mode + ".key"
	privKeyFileName := "privateKey_" + req.Mode + ".key"

	pubFile, err := zipWriter.Create(pubKeyFileName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if _, err := pubFile.Write(publicKey); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	privFile, err := zipWriter.Create(privKeyFileName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if _, err := privFile.Write(privateKey); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if err := zipWriter.Close(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	executionTime := time.Since(startTime)

	utils.SuccessResponse(c, "Key pair generated successfully", gin.H{
		"execution_time":     executionTime.Microseconds(),
		"generation_time":    endGenerate.Microseconds(),
		"memory_usage":       memUsage,
		"communication_size": communicationSize,
		"variant":            req.Mode,
	})
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
		utils.SuccessResponse(c, "Failed to Generate Key", gin.H{"error": err.Error()})
		return
	}

	publicKey, privateKey, err := dilithiumService.GenerateKeyPair()
	if err != nil {
		utils.SuccessResponse(c, "Failed to Generate Key", gin.H{"error": err.Error()})
		return
	}

	// Helper function to get the first and last 100 characters
	getKeyHash := func(key []byte) string {
		keyStr := string(key)
		if len(keyStr) <= 200 {
			// If the key length is less than or equal to 200, take the whole key
			keyStr = keyStr
		} else {
			// Otherwise, take the first 100 and last 100 characters
			keyStr = keyStr[:100] + keyStr[len(keyStr)-100:]
		}
		hash := sha256.Sum256([]byte(keyStr))
		return hex.EncodeToString(hash[:])
	}

	// Create hashes for public and private keys using the helper function
	pubKeyHashStr := getKeyHash(publicKey)
	privKeyHashStr := getKeyHash(privateKey)

	print(privKeyHashStr)

	// Save the key pair information in the database
	keyPair := &models.KeyPair{
		PrivateKey: privKeyHashStr,
		PublicKey:  pubKeyHashStr,
		Variant:    req.Mode,
	}

	_, err = ctrl.keyPairService.Create(c.Request.Context(), keyPair)
	if err != nil {
		utils.SuccessResponse(c, "Failed to save key pair to database", gin.H{"error": err.Error()})
		return
	}

	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Create filenames with mode in the ZIP
	pubKeyFileName := "publicKey_" + req.Mode + ".key"
	privKeyFileName := "privateKey_" + req.Mode + ".key"

	// Add the public key file to the zip
	pubFile, err := zipWriter.Create(pubKeyFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := pubFile.Write(publicKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add the private key file to the zip
	privFile, err := zipWriter.Create(privKeyFileName)
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

	// Serve the file with mode in the ZIP filename
	zipFileName := "keys_" + req.Mode + ".zip"
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+zipFileName)
	c.Writer.Header().Set("Content-Type", "application/zip")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(buf.Bytes())
}

func (ctrl *DilithiumController) SignMessage(c *gin.Context) {
	startExecution := time.Now()

	// Get private key file
	privateKeyFile, err := c.FormFile("privateKey")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Private key file is required", err)
		return
	}

	privateKeyFileContent, err := privateKeyFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open private key file", err)
		return
	}
	defer privateKeyFileContent.Close()

	privateKeyBytes, err := ioutil.ReadAll(privateKeyFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read private key file", err)
		return
	}

	// Helper function to get the first and last 100 characters
	getKeyHash := func(key []byte) string {
		keyStr := string(key)
		if len(keyStr) <= 200 {
			// If the key length is less than or equal to 200, take the whole key
			keyStr = keyStr
		} else {
			// Otherwise, take the first 100 and last 100 characters
			keyStr = keyStr[:100] + keyStr[len(keyStr)-100:]
		}
		hash := sha256.Sum256([]byte(keyStr))
		return hex.EncodeToString(hash[:])
	}

	// Hash the private key
	privKeyHashStr := getKeyHash(privateKeyBytes)

	// Find the key pair by hashed private key
	keyPairs, err := ctrl.keyPairService.FindByField(c.Request.Context(), "private_key", privKeyHashStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find key pair", err)
		return
	}

	if len(keyPairs) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Key pair not found", nil)
		return
	}

	// Get the variant (mode) from the key pair
	mode := keyPairs[0].Variant

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Get message file
	messageFile, err := c.FormFile("message")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Message file is required", err)
		return
	}

	messageFileContent, err := messageFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open message file", err)
		return
	}
	defer messageFileContent.Close()

	messageBytes, err := ioutil.ReadAll(messageFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read message file", err)
		return
	}

	// Capture memory usage before signing
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	startSign := time.Now()
	signature, err := dilithiumService.SignMessage(privateKeyBytes, messageBytes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to sign message", err)
		return
	}
	signTime := time.Since(startSign).Microseconds()

	// Capture memory usage after signing
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage during signing
	memUsage := memStatsAfter.Alloc - memStatsBefore.Alloc

	// Calculate communication size
	communicationSize := len(messageBytes) + len(signature)

	// Compute document hash
	documentHash := sha256.Sum256(messageBytes)
	documentHashStr := hex.EncodeToString(documentHash[:])

	// Create Signature object
	signatureModel := &models.Signature{
		KeyID:        keyPairs[0].ID,
		DocumentHash: documentHashStr,
		Signature:    hex.EncodeToString(signature),
	}

	// Save signature to database
	_, err = ctrl.signatureService.Create(c.Request.Context(), signatureModel)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save signature", err)
		return
	}
	executionTime := time.Since(startExecution).Microseconds()

	// Create response
	response := gin.H{
		"sign_time":                signTime,
		"execution_time":           executionTime,
		"memory_usage_bytes":       memUsage,
		"communication_size_bytes": communicationSize,
		"variant":                  mode,
	}

	// Send response
	utils.SuccessResponse(c, "Signature saved successfully", response)

	// Clear memory and communication size after sending the response
	memUsage = 0
	communicationSize = 0
	runtime.GC()
}

func (ctrl *DilithiumController) SignMessageUrl(c *gin.Context) {
	startExecution := time.Now()

	// Get private key file
	privateKeyFile, err := c.FormFile("privateKey")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Private key file is required", err)
		return
	}

	privateKeyFileContent, err := privateKeyFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open private key file", err)
		return
	}
	defer privateKeyFileContent.Close()

	privateKeyBytes, err := ioutil.ReadAll(privateKeyFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read private key file", err)
		return
	}

	// Helper function to get the first and last 100 characters
	getKeyHash := func(key []byte) string {
		keyStr := string(key)
		if len(keyStr) <= 200 {
			// If the key length is less than or equal to 200, take the whole key
			keyStr = keyStr
		} else {
			// Otherwise, take the first 100 and last 100 characters
			keyStr = keyStr[:100] + keyStr[len(keyStr)-100:]
		}
		hash := sha256.Sum256([]byte(keyStr))
		return hex.EncodeToString(hash[:])
	}

	// Hash the private key
	privKeyHashStr := getKeyHash(privateKeyBytes)

	// Find the key pair by hashed private key
	keyPairs, err := ctrl.keyPairService.FindByField(c.Request.Context(), "private_key", privKeyHashStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find key pair", err)
		return
	}

	print(keyPairs)

	if len(keyPairs) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Key pair not found", nil)
		return
	}

	// Get the variant (mode) from the key pair
	mode := keyPairs[0].Variant

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
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

	// Capture memory usage before signing
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	startSign := time.Now()
	signature, err := dilithiumService.SignMessage(privateKeyBytes, messageBytes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to sign message", err)
		return
	}
	signTime := time.Since(startSign).Microseconds()

	// Capture memory usage after signing
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage during signing
	memUsage := memStatsAfter.Alloc - memStatsBefore.Alloc

	// Calculate communication size
	communicationSize := len(messageBytes) + len(signature)

	// Compute document hash
	documentHash := sha256.Sum256(messageBytes)
	documentHashStr := hex.EncodeToString(documentHash[:])

	// Create Signature object
	signatureModel := &models.Signature{
		KeyID:        keyPairs[0].ID,
		DocumentHash: documentHashStr,
		Signature:    hex.EncodeToString(signature),
	}

	// Save signature to database
	_, err = ctrl.signatureService.Create(c.Request.Context(), signatureModel)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save signature", err)
		return
	}
	executionTime := time.Since(startExecution).Microseconds()

	utils.SuccessResponse(c, "Signature saved successfully", gin.H{
		"sign_time":                signTime,
		"execution_time":           executionTime,
		"memory_usage_bytes":       memUsage,
		"communication_size_bytes": communicationSize,
		"variant":                  mode,
	})
}

func (ctrl *DilithiumController) VerifySignature(c *gin.Context) {
	startExecution := time.Now()

	// Get public key file
	publicKeyFile, err := c.FormFile("publicKey")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Public key file is required", err)
		return
	}

	publicKeyFileContent, err := publicKeyFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open public key file", err)
		return
	}
	defer publicKeyFileContent.Close()

	publicKeyBytes, err := utils.ReadFileContent(publicKeyFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read public key file", err)
		return
	}

	// Get message file
	messageFile, err := c.FormFile("message")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Message file is required", err)
		return
	}

	messageFileContent, err := messageFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open message file", err)
		return
	}
	defer messageFileContent.Close()

	messageBytes, err := utils.ReadFileContent(messageFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read message file", err)
		return
	}

	// Hash the document
	documentHash := sha256.Sum256(messageBytes)
	documentHashStr := hex.EncodeToString(documentHash[:])

	// Helper function to get the first and last 100 characters
	getKeyHash := func(key []byte) string {
		keyStr := string(key)
		if len(keyStr) <= 200 {
			// If the key length is less than or equal to 200, take the whole key
			keyStr = keyStr
		} else {
			// Otherwise, take the first 100 and last 100 characters
			keyStr = keyStr[:100] + keyStr[len(keyStr)-100:]
		}
		hash := sha256.Sum256([]byte(keyStr))
		return hex.EncodeToString(hash[:])
	}

	// Hash the public key
	publicKeyHashStr := getKeyHash(publicKeyBytes)

	// Find the key pair by hashed public key
	keyPairs, err := ctrl.keyPairService.FindByField(c.Request.Context(), "public_key", publicKeyHashStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find key pair", err)
		return
	}

	if len(keyPairs) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Key pair not found", nil)
		return
	}

	// Get the key ID from the first key pair
	keyID := keyPairs[0].ID

	// Find the signature for the document hash and key ID
	signatures, err := ctrl.signatureService.FindByMessageHashAndKeyId(c.Request.Context(), documentHashStr, keyID)
	if len(signatures) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Signature not found", nil)
		return
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find signature", err)
		return
	}

	// Use the first signature found for verification
	signature := signatures[0]

	// Decode the signature from hex string to byte slice
	signatureBytes, err := hex.DecodeString(signature.Signature)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to decode signature", err)
		return
	}

	// Get the variant (mode) from the key pair
	mode := keyPairs[0].Variant

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Capture memory usage before verification
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	startVerification := time.Now()
	// Verify the signature
	valid, err := dilithiumService.VerifySignature(publicKeyBytes, messageBytes, signatureBytes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify signature", err)
		return
	}
	verificationTime := time.Since(startVerification).Microseconds()

	// Capture memory usage after verification
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage during verification
	memUsage := memStatsAfter.Alloc - memStatsBefore.Alloc

	// Calculate communication size
	communicationSize := len(publicKeyBytes) + len(messageBytes) + len(signatureBytes)

	executionTime := time.Since(startExecution).Microseconds()

	// Return JSON response with validity
	utils.SuccessResponse(c, "Signature verification result", gin.H{
		"valid":                    valid,
		"verification_time":        verificationTime,
		"execution_time":           executionTime,
		"memory_usage_bytes":       memUsage,
		"communication_size_bytes": communicationSize,
		"variant":                  mode,
	})
}

func (ctrl *DilithiumController) VerifySignatureUrl(c *gin.Context) {
	startExecution := time.Now()

	// Get public key file
	publicKeyFile, err := c.FormFile("publicKey")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Public key file is required", err)
		return
	}

	publicKeyFileContent, err := publicKeyFile.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open public key file", err)
		return
	}
	defer publicKeyFileContent.Close()

	publicKeyBytes, err := utils.ReadFileContent(publicKeyFileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read public key file", err)
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

	// Hash the document
	documentHash := sha256.Sum256(messageBytes)
	documentHashStr := hex.EncodeToString(documentHash[:])

	// Helper function to get the first and last 100 characters
	getKeyHash := func(key []byte) string {
		keyStr := string(key)
		if len(keyStr) <= 200 {
			// If the key length is less than or equal to 200, take the whole key
			keyStr = keyStr
		} else {
			// Otherwise, take the first 100 and last 100 characters
			keyStr = keyStr[:100] + keyStr[len(keyStr)-100:]
		}
		hash := sha256.Sum256([]byte(keyStr))
		return hex.EncodeToString(hash[:])
	}

	// Hash the public key
	publicKeyHashStr := getKeyHash(publicKeyBytes)

	// Find the key pair by hashed public key
	keyPairs, err := ctrl.keyPairService.FindByField(c.Request.Context(), "public_key", publicKeyHashStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find key pair", err)
		return
	}

	if len(keyPairs) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Key pair not found", nil)
		return
	}

	// Get the key ID from the first key pair
	keyID := keyPairs[0].ID

	// Find the signature for the document hash and key ID
	signatures, err := ctrl.signatureService.FindByMessageHashAndKeyId(c.Request.Context(), documentHashStr, keyID)
	if len(signatures) == 0 {
		utils.SuccessResponse(c, "Signature not found", nil)
		return
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find signature", err)
		return
	}

	// Use the first signature found for verification
	signature := signatures[0]

	// Decode the signature from hex string to byte slice
	signatureBytes, err := hex.DecodeString(signature.Signature)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to decode signature", err)
		return
	}

	// Get the variant (mode) from the key pair
	mode := keyPairs[0].Variant

	dilithiumService, err := ctrl.serviceFactory(mode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Capture memory usage before verification
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	startVerification := time.Now()
	// Verify the signature
	valid, err := dilithiumService.VerifySignature(publicKeyBytes, messageBytes, signatureBytes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify signature", err)
		return
	}
	verificationTime := time.Since(startVerification).Microseconds()

	// Capture memory usage after verification
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage during verification
	memUsage := memStatsAfter.Alloc - memStatsBefore.Alloc

	// Calculate communication size
	communicationSize := len(publicKeyBytes) + len(messageBytes) + len(signatureBytes)

	executionTime := time.Since(startExecution).Microseconds()

	// Return JSON response with validity
	utils.SuccessResponse(c, "Signature verification result", gin.H{
		"valid":                    valid,
		"verification_time":        verificationTime,
		"execution_time":           executionTime,
		"memory_usage_bytes":       memUsage,
		"communication_size_bytes": communicationSize,
		"variant":                  mode,
	})
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
