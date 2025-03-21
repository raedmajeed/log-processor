package services

import (
	"LOGProcessor/log-mainService/tasks"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

/******************************************************************************
* FUNCTION:        HandleUploadFileToQueue
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func HandleAsyncTaskMethod(ctx context.Context, t *asynq.Task) error {
	fmt.Println("PAYLOAD RECEIVED")
	payload := t.Payload()
	var pay tasks.LogProcessPayload
	json.Unmarshal(payload, &pay)

	queueName := "high"
	if pay.FileSizeBytes > 1073741824 {
		queueName = "low"
	}

	fmt.Println("PAYLOAD RECEIVED", pay, queueName)
	// services.TesttingThis()
	//* This function handles the workers responsible to stream the logs
	return nil
}
