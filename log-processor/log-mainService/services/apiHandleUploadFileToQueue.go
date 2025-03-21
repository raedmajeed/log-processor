/**************************************************************************
 * File       	   : apiHandleUploadFileToQueue.go
 * DESCRIPTION     : This file contains functions that uploads files to
 *									 supabase storage and also enqueu to Async queue
 * DATE            : 16-March-2025
 **************************************************************************/

package services

import (
	// "LOGProcessor/log-mainService/tasks"
	"LOGProcessor/log-mainService/tasks"
	"LOGProcessor/shared/types"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	storage_go "github.com/supabase-community/storage-go"
)

/******************************************************************************
* FUNCTION:        HandleUploadFileToQueue
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func HandleUploadFileToQueue(ctx *gin.Context) {
	var (
		err        error
		fileHeader *multipart.FileHeader
		file       multipart.File
		uploadRsp  storage_go.FileUploadResponse
		// token      string
		// fileName   string
		filePath string
		fileSize int64
	)

	fileHeader, err = ctx.FormFile("log-file")
	if err != nil {
		SendResponse(ctx, http.StatusBadRequest, "failed to read file", "", 0)
		return
	}

	// fileName = fileHeader.Filename
	file, err = fileHeader.Open()
	if err != nil {
		SendResponse(ctx, http.StatusBadRequest, "failed to read file", "", 0)
		return
	}
	defer file.Close()
	fileSize = fileHeader.Size

	// token, err = extractToken(ctx)
	// if err != nil {
	// 	SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
	// 	return
	// }

	// uploadRsp, err = uploadFileToSupeBaseStorage(token, fileName, file)
	// if err != nil {
	// 	SendResponse(ctx, http.StatusBadRequest, "internal server error", "", 0)
	// 	return
	// }
	uploadRsp.Key = "working/superb"
	filePath = uploadRsp.Key
	fileId := "100"

	task, _ := tasks.NewLogProcessTask(fileId, filePath, fileSize)
	taskInfo, err := types.AsynqClient.AsynqClient.Enqueue(task)
	if err != nil {
		fmt.Println("errorrro", err)
	}

	fmt.Println("STATE", taskInfo.State.String())

	fmt.Println("TASK INFO", taskInfo.ID)

	SendResponse(ctx, http.StatusOK, "File uploaded successfully", filePath, 1)
}
