/**************************************************************************
 * File       	   : apiHandleGetAggregateStats.go
 * DESCRIPTION     : This file contains functions that gets the file stats
 * DATE            : 22-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/shared/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
)

func HandleGetAggregatedTasks(ctx *gin.Context) {
	var (
		err          error
		query        string
		whereEleList []interface{}
		result       []map[string]interface{}
		userId       string
	)

	userId, err = extractToken(ctx, "user_id")
	if err != nil {
		log.Errorf("failed to get user_id from context; err: ", err)
		SendResponse(ctx, http.StatusUnauthorized, "internal server error", "", 0)
		return
	}

	query = `
	SELECT * FROM file_stats WHERE user_id = $1`

	whereEleList = append(whereEleList, userId)
	result, err = db.GetDataFromDB(query, whereEleList)
	if err != nil {
		log.Errorf("failed to get data from db; err: ", err)
		SendResponse(ctx, http.StatusBadRequest, "internal server error", nil, 0)
		return
	}
	SendResponse(ctx, http.StatusOK, "aggregated stats retrieved succesfully", result, int64(len(result)))
}
