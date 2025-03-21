/**************************************************************************
 * File       	   : apiSupeBaseStorageHandlers.go
 * DESCRIPTION     : This file contains functions that uploads and downloads
 *									 files from supebase storage
 * DATE            : 16-March-2025
 **************************************************************************/

package services

import (
	"LOGProcessor/shared/types"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	storage "github.com/supabase-community/storage-go"
)

/******************************************************************************
* FUNCTION:        uploadFileToSupeBaseStorage
*
* DESCRIPTION:     This function is used to upload file to SupeBase
* INPUT:
* RETURNS:         string, err
******************************************************************************/
func uploadFileToSupeBaseStorage(authToken, fileName string, file multipart.File) (storage.FileUploadResponse, error) {
	client := ConfigSupeBaseStorageClient(authToken)
	resp, err := client.UploadFile(types.CmnGlblCfg.SUPEBASE_BUCKET, fileName, file)

	if err != nil {
		return storage.FileUploadResponse{}, err
	}

	return resp, nil
}

/******************************************************************************
* FUNCTION:        downloadFileFromSupeBaseStorage
*
* DESCRIPTION:     This function is used to download file from SupeBase
* INPUT:
* RETURNS:         io.ReadCloser, err
******************************************************************************/
func downloadFileFromSupeBaseStorage(filePath string) (io.ReadCloser, error) {
	token, err := JwtTokenCreatorForSupebaseStorage(filePath)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s/object/sign/%s?token=%s",
		types.CmnGlblCfg.SUPEBASE_API, types.CmnGlblCfg.SUPEBASE_STORAGE_BASE, filePath, token)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download file, status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
