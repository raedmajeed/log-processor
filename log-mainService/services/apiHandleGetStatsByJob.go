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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
)

/******************************************************************************
* FUNCTION:        HandleGetStatsByJobId
*
* DESCRIPTION:     This function gets log_stats by job id
*									 The thing to not is, here the pagination should happen incrementally
*									 the reason being paginating with offset can be very expensive and time consuming
*									 for very large files, which is expected from a logging service.
*									 Hence to optimize querying I take an approach by filtering with
*									 the last id, thereby resulting in a faster output
*
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func HandleGetStatsByJobId(ctx *gin.Context) {
	PanicRecovery("HandleGetStatsByJobId")

	var (
		jobId        string
		err          error
		query        string
		userId       string
		whereEleList []interface{}
		result       []map[string]interface{}
	)

	jobId = ctx.Param("jobId")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	lastIdStr := ctx.DefaultQuery("lastId", "0")

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		SendResponse(ctx, http.StatusBadRequest, "invalid pageSize", nil, 0)
		return
	}

	lastId, err := strconv.ParseInt(lastIdStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse lastId; err: %v", err)
		SendResponse(ctx, http.StatusBadRequest, "invalid lastId", nil, 0)
		return
	}

	userId, err = extractToken(ctx, "user_id")
	if err != nil {
		log.Errorf("failed to get user_id from context; err: ", err)
		SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
		return
	}

	query = `
	SELECT l.* FROM log_stats l
	JOIN file_stats f ON l.file_id = f.file_id
	WHERE f.job_id = $1 AND f.user_id = $2 AND l.file_id > $3
	ORDER BY l.file_id ASC
	LIMIT $4`

	whereEleList = append(whereEleList, jobId, userId, lastId, pageSize)
	result, err = db.GetDataFromDB(query, whereEleList)
	if err != nil {
		log.Errorf("failed to get data from db; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", nil, 0)
		return
	}

	var nextLastId int64
	if len(result) > 0 {
		nextLastId = result[len(result)-1]["file_id"].(int64)
	}

	responseData := map[string]interface{}{
		"data":       result,
		"nextLastId": nextLastId,
		"pageSize":   pageSize,
	}

	SendResponse(ctx, http.StatusOK, "log stats retrieved succesfully", responseData, int64(len(result)))
}
