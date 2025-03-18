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
	FileId   string
	FilePath string
}

/******************************************************************************
* FUNCTION:        NewLogProcessTask
*
* DESCRIPTION:     This function is used to create new asynq task
* INPUT:					 fileId, filePath
* RETURNS:         *asynq.Task, error
******************************************************************************/
func NewLogProcessTask(fileId, filePath string) (*asynq.Task, error) {
	var (
		err error
	)

	payload, err := json.Marshal(LogProcessPayload{
		FileId:   fileId,
		FilePath: filePath,
	})

	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeLogProcess, payload), nil
}

// func HandleLogProcessTask(ctx context.Context, t *asynq.Task) error {

// }
