package services

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func TestService(ctx *gin.Context) {
	fmt.Println("WORKING GREAT")
}
