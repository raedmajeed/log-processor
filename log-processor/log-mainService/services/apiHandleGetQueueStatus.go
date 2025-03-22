/**************************************************************************
 * File       	   : apiHandleGetQueueStatus.go
 * DESCRIPTION     : This file contains functions that gets the current
 *                   status of the asynq queue
 * DATE            : 22-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/log-mainService/tasks"
	"LOGProcessor/shared/types"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type QueueResp struct {
	HighQueue QueueInfo
	LowQueue  QueueInfo
	Total     QueueInfo
}

type QueueInfo struct {
	Pending   int `json:"pending"`
	Active    int `json:"active"`
	Scheduled int `json:"scheduled"`
	Retry     int `json:"retry"`
	Processed int `json:"processed"`
	Failed    int `json:"failed"`
}

func HandleGetQueueCurrentStatus(ctx *gin.Context) {
	var (
		err       error
		inspector *asynq.Inspector
	)
	inspector = asynq.NewInspector(asynq.RedisClientOpt{
		Addr: types.CmnGlblCfg.REDIS_ADDR,
	})

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	highQueueStats, err := getQueueStats(timeoutCtx, inspector, "high")
	if err != nil {
		SendResponse(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get high queue stats: %v", err), "", 0)
		return
	}

	lowQueueStats, err := getQueueStats(timeoutCtx, inspector, "low")
	if err != nil {
		SendResponse(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to get low queue stats: %v", err), "", 0)
		return
	}

	totalStats := QueueInfo{
		Pending:   highQueueStats.Pending + lowQueueStats.Pending,
		Active:    highQueueStats.Active + lowQueueStats.Active,
		Scheduled: highQueueStats.Scheduled + lowQueueStats.Scheduled,
		Retry:     highQueueStats.Retry + lowQueueStats.Retry,
		Processed: highQueueStats.Processed + lowQueueStats.Processed,
		Failed:    highQueueStats.Failed + lowQueueStats.Failed,
	}

	queueStatus := QueueResp{
		HighQueue: highQueueStats,
		LowQueue:  lowQueueStats,
		Total:     totalStats,
	}

	activeTasks, err := getActiveTasks(inspector)
	if err == nil && len(activeTasks) > 0 {
		responseData := map[string]interface{}{
			"queue_status": queueStatus,
			"active_tasks": activeTasks,
		}
		SendResponse(ctx, http.StatusOK, "Queue status fetched successfully", responseData, 1)
	} else {
		SendResponse(ctx, http.StatusOK, "Queue status fetched successfully", queueStatus, 1)
	}

}

/******************************************************************************
* FUNCTION:        getQueueStats
*
* DESCRIPTION:     Helper function to get details of queue stats
* INPUT:           inspector
* RETURNS:         QueueInfo, error
******************************************************************************/
func getQueueStats(c context.Context, inspector *asynq.Inspector, qName string) (QueueInfo, error) {
	qInfo, err := inspector.GetQueueInfo(qName)
	if err != nil {
		return QueueInfo{}, err
	}

	return QueueInfo{
		Pending:   qInfo.Pending,
		Active:    qInfo.Active,
		Scheduled: qInfo.Scheduled,
		Retry:     qInfo.Retry,
		Processed: qInfo.Processed,
		Failed:    qInfo.Failed,
	}, nil

}

/******************************************************************************
* FUNCTION:        getActiveTasks
*
* DESCRIPTION:     Helper function to get details of active tasks
* INPUT:           inspector
* RETURNS:         []map[string]interface{}, error
******************************************************************************/
func getActiveTasks(inspector *asynq.Inspector) ([]map[string]interface{}, error) {
	highActiveTasks, err1 := inspector.ListActiveTasks("high", asynq.PageSize(10))
	lowActiveTasks, err2 := inspector.ListActiveTasks("low", asynq.PageSize(10))

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf("failed to get active tasks")
	}

	allTasks := append(highActiveTasks, lowActiveTasks...)
	result := make([]map[string]interface{}, 0, len(allTasks))

	for _, task := range allTasks {
		taskMap := map[string]interface{}{
			"id":        task.ID,
			"type":      task.Type,
			"queue":     task.Queue,
			"max_retry": task.MaxRetry,
			"retried":   task.Retried,
		}

		if task.Type == tasks.TypeLogProcess {
			var payload tasks.LogProcessPayload
			if err := json.Unmarshal(task.Payload, &payload); err == nil {
				taskMap["payload"] = payload
			}
		}

		result = append(result, taskMap)
	}

	return result, nil
}
