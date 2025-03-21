/**************************************************************************
 * File       	   : tasks.go
 * DESCRIPTION     : This file contains functions that helps in creating
 *									 asynq tasks
 * DATE            : 1-March-2025
 **************************************************************************/

package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeLogProcess = "log:process"
)

type LogProcessPayload struct {
	FileId        int64
	FilePath      string
	FileSizeBytes int64
	UserId        string
}

/******************************************************************************
* FUNCTION:        NewLogProcessTask
*
* DESCRIPTION:     This function is used to create new asynq task
* INPUT:					 fileId, filePath
* RETURNS:         *asynq.Task, error
******************************************************************************/
func NewLogProcessTask(filePath, userId string, fileId int64, fileSizeBytes int64) (*asynq.Task, error) {
	var (
		err       error
		queueName string
		options   []asynq.Option
	)

	if fileSizeBytes > 1073741824 {
		queueName = "low"
	} else {
		queueName = "high"
	}

	payload, err := json.Marshal(LogProcessPayload{
		FileId:        fileId,
		FilePath:      filePath,
		FileSizeBytes: fileSizeBytes,
		UserId:        userId,
	})

	if err != nil {
		return nil, err
	}

	options = []asynq.Option{
		asynq.Queue(queueName),
		asynq.MaxRetry(3),
	}

	return asynq.NewTask(TypeLogProcess, payload, options...), nil
}

// func HandleLogProcessTask(ctx context.Context, t *asynq.Task) error {

// }
