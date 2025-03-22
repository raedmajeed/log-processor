/**************************************************************************
 * File       	   : apiHandleGetStatsByJob.go
 * DESCRIPTION     : This file contains functions that gets the log stats
 *                   by job id
 * DATE            : 22-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/shared/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
)

func HandleGetStatsByJobId(ctx *gin.Context) {
	var (
		jobId        string
		err          error
		query        string
		userId       string
		whereEleList []interface{}
		result       []map[string]interface{}
	)

	jobId = ctx.Param("jobId")

	userId, err = extractToken(ctx, "user_id")
	if err != nil {
		log.Errorf("failed to get user_id from context; err: ", err)
		SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
		return
	}

	query = `
	SELECT * FROM log_stats l JOIN file_stats f ON l.file_id = f.file_id
	WHERE f.job_id = $1 AND f.user_id = $2`

	whereEleList = append(whereEleList, jobId, userId)
	result, err = db.GetDataFromDB(query, whereEleList)
	if err != nil {
		log.Errorf("failed to get data from db; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", nil, 0)
		return
	}
	SendResponse(ctx, http.StatusOK, "log stats retrieved succesfully", result, int64(len(result)))
}
