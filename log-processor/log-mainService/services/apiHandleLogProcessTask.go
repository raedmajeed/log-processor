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
	fmt.Println("PAYLOAD RECEIVED", pay)
	// services.TesttingThis()
	//* This function handles the workers responsible to stream the logs
	return nil
}
