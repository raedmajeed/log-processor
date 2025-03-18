/**************************************************************************
 * File       	   : apiTokenBuilder.go
 * DESCRIPTION     : This file contains Helper functions
 * DATE            : 16-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/shared/types"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	storage "github.com/supabase-community/storage-go"
)

/******************************************************************************
* FUNCTION:        JwtTokenCreatorForSupebaseStorage
*
* DESCRIPTION:     This function is used to create the JWT token that is necessary
* 								 to download file from Supebase
* INPUT:           filepath
* RETURNS:         string, err
******************************************************************************/
func JwtTokenCreatorForSupebaseStorage(filePath string) (token string, err error) {
	claims := jwt.MapClaims{
		"url": filePath,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = unsignedToken.SignedString(types.CmnGlblCfg.JWT_SECRET)

	if err != nil {
		return "", err
	}
	return token, nil
}

/******************************************************************************
* FUNCTION:        ConfigSupeBaseStorageClient
* DESCRIPTION:     Function to set supebase storage client
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func ConfigSupeBaseStorageClient(authToken string) *storage.Client {
	headers := map[string]string{
		"authorization": "bearer " + authToken,
	}
	client := storage.NewClient(types.CmnGlblCfg.SUPEBASE_API+types.CmnGlblCfg.SUPEBASE_STORAGE_BASE, types.CmnGlblCfg.SUPEBASE_API_KEY, headers)
	return client
}

/******************************************************************************
* FUNCTION:        extractToken
*
* DESCRIPTION:     Extracts the token from the Gin context.
* INPUT:           ctx *gin.Context
* RETURNS:         string (token), error (if missing)
******************************************************************************/
func extractToken(ctx *gin.Context) (string, error) {
	tokenInt, exists := ctx.Get("token")
	if !exists {
		return "", fmt.Errorf("token not found")
	}

	token, ok := tokenInt.(string)
	if !ok {
		return "", fmt.Errorf("invalid token format")
	}

	return token, nil
}

/******************************************************************************
 * FUNCTION: SendResponse
 * DESCRIPTION: Generic function to send JSON responses with consistent format
 * INPUT: ctx (gin context), status code, message, data map, record count
 * RETURNS: none
 ******************************************************************************/
func SendResponse(ctx *gin.Context, statusCode int, message string, data types.Data, dbRecordCount int64) {

	response := types.ApiResponse{
		Status:     statusCode,
		Message:    message,
		DbRecCount: dbRecordCount,
		Data:       data,
	}

	ctx.JSON(statusCode, response)
}
