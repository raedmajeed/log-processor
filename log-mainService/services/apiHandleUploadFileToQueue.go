/**************************************************************************
 * File       	   : apiHandleUploadFileToQueue.go
 * DESCRIPTION     : This file contains functions that uploads files to
 *									 supabase storage and also enqueu to Async queue
 * DATE            : 16-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/log-mainService/tasks"
	"LOGProcessor/shared/db"
	"LOGProcessor/shared/types"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
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
	defer PanicRecovery("HandleUploadFileToQueue")

	var (
		err        error
		fileHeader *multipart.FileHeader
		file       multipart.File
		uploadRsp  storage_go.FileUploadResponse
		token      string
		fileName   string
		filePath   string
		fileSize   int64
		userId     string
		fileId     int64
	)

	fileHeader, err = ctx.FormFile("log-file")
	if err != nil {
		SendResponse(ctx, http.StatusBadRequest, "failed to read file", "", 0)
		return
	}

	fileName = fileHeader.Filename
	fileSize = fileHeader.Size
	file, err = fileHeader.Open()
	if err != nil {
		log.Errorf("failed to read file; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "failed to read file", "", 0)
		return
	}
	defer file.Close()

	token, err = extractToken(ctx, "token")
	if err != nil {
		log.Errorf("failed to get token from context; err: ", err)
		SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
		return
	}
	userId, err = extractToken(ctx, "user_id")
	if err != nil {
		log.Errorf("failed to get user_id from context; err: ", err)
		SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
		return
	}

	uploadRsp, err = uploadFileToSupeBaseStorage(token, fileName, file)
	if err != nil {
		if strings.Contains(err.Error(), "The resource already exists") {
			SendResponse(ctx, http.StatusBadRequest, "file name already exists", "", 0)
			return
		}
		log.Errorf("failed to upload to supabase; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", "", 0)
		return
	}

	filePath = uploadRsp.Key
	data := map[string]interface{}{
		"file_name":    fileName,
		"file_size_mb": float64(fileSize) / (1024 * 1024),
		"status":       "pending",
		"created_at":   time.Now(),
		"file_path":    filePath,
		"user_id":      userId,
	}

	dbConn := types.Db.DbConn
	tx, _ := dbConn.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Errorf("panic occurred: %v", p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	fileId, err = db.InsertAndReturnID(tx, "file_stats", data)
	if err != nil {
		log.Errorf("failed to insert into file_stats; err: %v", err)
		SendResponse(ctx, http.StatusInternalServerError, "internal server error", "", 0)
		return
	}

	task, _ := tasks.NewLogProcessTask(filePath, userId, fileId, fileSize)
	taskInfo, err := types.AsynqClient.AsynqClient.Enqueue(task)
	if err != nil {
		log.Errorf("failed to create task; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", "", 0)
		return
	}

	_, err = db.UpdateDataInDB(tx, "UPDATE file_stats SET job_id = $1 WHERE file_id = $2", []interface{}{taskInfo.ID, fileId})
	if err != nil {
		log.Errorf("failed to update job_id; err: %v", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", "", 0)
		return
	}

	data["file_id"] = fileId
	data["job_id"] = taskInfo.ID
	BroadcastMessage(data, "log-table-update", userId)
	SendResponse(ctx, http.StatusOK, "File uploaded successfully", filePath, 1)
}
