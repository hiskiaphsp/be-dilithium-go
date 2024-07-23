package utils

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func CreatedResponse(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusCreated, Response{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(ctx *gin.Context, status int, message string, err error) {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	ctx.JSON(status, Response{
		Status:  status,
		Message: message,
		Error:   errorMessage,
	})
}

// ReadFileContent reads the content of an io.Reader and returns it as a byte slice.
func ReadFileContent(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}
