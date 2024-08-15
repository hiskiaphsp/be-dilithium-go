# Backend Dilithium Using Go (Gin)

## Overview

This project is a backend implementation using Go and the Gin framework. It focuses on using the Dilithium post-quantum digital signature algorithm for secure communications.
LIB: 
https://pkg.go.dev/github.com/cloudflare/circl/sign/dilithium

## Requirements

- Go 1.22.4

## Modules

This project uses the following Go modules:

```go
module be-dilithium

go 1.22.4

require (
    github.com/cloudflare/circl/sign/dilithium/mode2
    github.com/gin-gonic/gin v1.10.0 // indirect
    github.com/go-sql-driver/mysql v1.8.1 // indirect
    github.com/joho/godotenv v1.5.1 // indirect
    gorm.io/driver/mysql v1.5.7 // indirect
    gorm.io/gorm v1.25.11 // indirect
)

# RUN PROJECT
``go run main.go```

## API Endpoints

### Key Pairs

- **Generate Key Pair**
  - **URL**: `/generate-keypair`
  - **Method**: `POST`
  - **Description**: Generates a new key pair.

- **Generate Key Pair with Time**
  - **URL**: `/generate-keypair-time`
  - **Method**: `POST`
  - **Description**: Generates a new key pair and returns the key generation time.

### Signatures

- **Sign Message**
  - **URL**: `/sign-message`
  - **Method**: `POST`
  - **Description**: Signs a message.

- **Sign Message by URL**
  - **URL**: `/sign-message-url`
  - **Method**: `POST`
  - **Description**: Signs a message by providing a URL.

- **Verify Signature**
  - **URL**: `/verify-signature`
  - **Method**: `POST`
  - **Description**: Verifies a signature.

- **Verify Signature by URL**
  - **URL**: `/verify-signature-url`
  - **Method**: `POST`
  - **Description**: Verifies a signature by providing a URL.



