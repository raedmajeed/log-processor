package services

import (
	"LOGProcessor/log-mainService/tasks"
	"LOGProcessor/shared/db"
	"LOGProcessor/shared/types"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

const (
	MAX_CHUNKS = 1
)

type FileChunk struct {
	StartOffset int64
	EndOffset   int64
	ChunkIndex  int
}

type LogEntry struct {
	Timestamp       time.Time
	LogLevel        string
	Message         string
	KeywordDetected string
	IP              string
	FileID          int64
}

type KeywordStats map[string]int

type LogStats struct {
	LogEntries    []LogEntry
	FileSize      int64
	FileName      string
	FilePath      string
	UserID        string
	JobID         string
	KeywordCounts KeywordStats
	ErrorCount    int
}

var logLineRegex = regexp.MustCompile(`\[(.*?)\]\s+(\w+)\s+(.*)`)
var ipRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
var keywordList = ConvertToKeywordList(types.CmnGlblCfg.KEYWORD_CONFIG)


/******************************************************************************
* FUNCTION:        HandleUploadFileToQueue
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func HandleAsyncTaskMethod(ctx context.Context, t *asynq.Task) error {
	var (
		err           error
		logStats      *LogStats
		fileSizeBytes int64
	)
	payload := t.Payload()
	var pay tasks.LogProcessPayload
	json.Unmarshal(payload, &pay)
	startTime := time.Now()

	dbConn := types.Db.DbConn
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		updateFileStats(pay.FileId, "Failed", startTime, 0, fmt.Sprintf("failed to start transaction: %v", err))
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			errorMsg := fmt.Sprintf("panic occurred during processing: %v", r)
			updateFileStats(pay.FileId, "Failed", startTime, 0, errorMsg)
			fmt.Println(errorMsg)
		} else if err != nil {
			tx.Rollback()
			updateFileStats(pay.FileId, "Failed", startTime, 0, fmt.Sprintf("error during processing: %v", err))
			fmt.Println("Transaction rolled back due to error:", err)
		}
	}()

	if fileSizeBytes < 1073741824 {
		logStats, err = processLargeLogFile(ctx, pay.FilePath, pay.FileId, pay.FileSizeBytes)
	} else {
		logStats, err = processLogFile(ctx, pay.FilePath, pay.FileId)
		if err != nil {
			return fmt.Errorf("error processing log file: %v", err)
		}
	}

	err = insertLogEntries(ctx, logStats.LogEntries)
	if err != nil {
		return fmt.Errorf("error inserting log entries: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	keywordJSON, _ := json.Marshal(logStats.KeywordCounts)
	updateFileStats(pay.FileId, "Completed", startTime, logStats.ErrorCount, "", string(keywordJSON))

	fmt.Println("Log processing completed successfully", logStats)
	return nil
}

/******************************************************************************
* FUNCTION:        processLogFile
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func processLogFile(ctx context.Context, filePath string, fileID int64) (*LogStats, error) {

	fileContent, err := downloadFileFromSupeBaseStorage(filePath)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %v", err)
	}

	tempFile, err := os.CreateTemp("", "log-processing-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, fileContent)
	if err != nil {
		return nil, fmt.Errorf("error writing to temp file: %v", err)
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error resetting file pointer: %v", err)
	}

	scanner := bufio.NewScanner(tempFile)
	logEntries := []LogEntry{}
	keywordCounts := make(KeywordStats)
	errorCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		entry, ok := parseLogLine(line, fileID)
		if ok {
			logEntries = append(logEntries, entry)
		}
		if entry.KeywordDetected != "" {
			keywordCounts[entry.KeywordDetected]++
			errorCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning log file: %v", err)
	}

	logStats := &LogStats{
		LogEntries:    logEntries,
		KeywordCounts: keywordCounts,
		ErrorCount:    errorCount,
	}

	return logStats, nil
}

/******************************************************************************
* FUNCTION:        parseLogLine
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func parseLogLine(line string, fileID int64) (LogEntry, bool) {
	matches := logLineRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return LogEntry{}, false
	}

	timestamp, err := time.Parse(time.RFC3339, matches[1])
	if err != nil {
		timestamp, err = time.Parse("2006-01-02 15:04:05", matches[1])
		if err != nil {
			return LogEntry{}, false
		}
	}

	logLevel := strings.ToUpper(matches[2])
	message := matches[3]

	var jsonPayload map[string]interface{}
	ip := ""

	jsonStart := strings.Index(message, "{")
	if jsonStart != -1 {
		jsonString := message[jsonStart:]
		err := json.Unmarshal([]byte(jsonString), &jsonPayload)
		if err == nil {
			message = strings.TrimSpace(message[:jsonStart])
			if ipValue, ok := jsonPayload["ip"]; ok {
				ip = fmt.Sprintf("%v", ipValue)
			}
		}
	}

	if ip == "" {
		ipMatches := ipRegex.FindStringSubmatch(message)
		if len(ipMatches) > 0 {
			ip = ipMatches[0]
		}
	}

	keywordDetected := ""
	messageLower := strings.ToLower(message)
	for _, keyword := range keywordList {
		if strings.Contains(messageLower, keyword) {
			keywordDetected = keyword
			break
		}
	}

	return LogEntry{
		Timestamp:       timestamp,
		LogLevel:        logLevel,
		Message:         message,
		KeywordDetected: keywordDetected,
		IP:              ip,
		FileID:          fileID,
	}, true
}

/******************************************************************************
* FUNCTION:        insertLogEntries
*
* DESCRIPTION:     This function is used to upload file to SupeBase & enqueu
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func insertLogEntries(ctx context.Context, entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	logData := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		logData = append(logData, map[string]interface{}{
			"file_id":          entry.FileID,
			"err_timestamp":    entry.Timestamp,
			"log_level":        entry.LogLevel,
			"err_mssg":         entry.Message,
			"keyword_detected": entry.KeywordDetected,
			"ip":               entry.IP,
			"created_at":       time.Now(),
		})
	}

	batchSize := 1000
	for i := 0; i < len(logData); i += batchSize {
		end := i + batchSize
		if end > len(logData) {
			end = len(logData)
		}

		batch := logData[i:end]
		err := db.AddMultipleRecordInDB("log_stats", batch)
		if err != nil {
			return fmt.Errorf("error inserting log batch: %v", err)
		}
	}

	return nil
}

/******************************************************************************
* FUNCTION:        processLargeLogFile
*
* DESCRIPTION:     Process large log files by breaking into chunks
* INPUT:           Context, file path, ID, size, user ID, job ID
* RETURNS:         LogStats, error
******************************************************************************/
func processLargeLogFile(ctx context.Context, filePath string, fileID int64, fileSize int64) (*LogStats, error) {

	fileContent, err := downloadFileFromSupeBaseStorage(filePath)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %v", err)
	}

	tempFile, err := os.CreateTemp("", "log-processing-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, fileContent)
	if err != nil {
		return nil, fmt.Errorf("error writing to temp file: %v", err)
	}

	chunkSize := fileSize / MAX_CHUNKS
	var chunks []FileChunk

	for i := 0; i < MAX_CHUNKS; i++ {
		startOffset := int64(i) * chunkSize
		endOffset := startOffset + chunkSize
		if i == MAX_CHUNKS-1 {
			endOffset = fileSize
		}
		chunks = append(chunks, FileChunk{
			StartOffset: startOffset,
			EndOffset:   endOffset,
			ChunkIndex:  i,
		})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allLogEntries := []LogEntry{}
	chunkErrors := make([]error, MAX_CHUNKS)
	finalKeywordStats := make(KeywordStats)
	finalErrCount := 0

	for i, chunk := range chunks {
		wg.Add(1)
		go func(c FileChunk, index int) {
			defer wg.Done()

			chunkEntries, keywordStats, errCount, err := processFileChunk(tempFile, c, fileID)
			if err != nil {
				chunkErrors[index] = err
				return
			}

			mu.Lock()
			for k, _ := range keywordStats {
				if _, ok := finalKeywordStats[k]; !ok {
					finalKeywordStats[k] = keywordStats[k]
				} else {
					finalKeywordStats[k] += keywordStats[k]
				}
			}
			finalErrCount += errCount
			allLogEntries = append(allLogEntries, chunkEntries...)
			mu.Unlock()
		}(chunk, i)
	}

	wg.Wait()

	for i, err := range chunkErrors {
		if err != nil {
			return nil, fmt.Errorf("error processing chunk %d: %v", i, err)
		}
	}

	return &LogStats{
		LogEntries:    allLogEntries,
		KeywordCounts: finalKeywordStats,
		ErrorCount:    finalErrCount,
	}, nil
}

/******************************************************************************
* FUNCTION:        processFileChunk
*
* DESCRIPTION:     Process a single chunk of a log file
* INPUT:           File, chunk info, file ID
* RETURNS:         Log entries, error
******************************************************************************/
func processFileChunk(file *os.File, chunk FileChunk, fileID int64) ([]LogEntry, KeywordStats, int, error) {
	_, err := file.Seek(chunk.StartOffset, 0)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error seeking to chunk start: %v", err)
	}

	if chunk.StartOffset > 0 {
		reader := bufio.NewReader(file)
		_, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, nil, 0, fmt.Errorf("error finding line boundary: %v", err)
		}
	}

	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error getting current position: %v", err)
	}

	chunkSize := chunk.EndOffset - currentPos
	if chunkSize <= 0 {
		return []LogEntry{}, nil, 0, nil
	}

	logEntries := []LogEntry{}
	scanner := bufio.NewScanner(io.LimitReader(file, chunkSize))
	errorCount := 0
	keywordCounts := make(KeywordStats)

	for scanner.Scan() {
		line := scanner.Text()
		entry, ok := parseLogLine(line, fileID)
		if ok {
			logEntries = append(logEntries, entry)
		}
		if entry.KeywordDetected != "" {
			keywordCounts[entry.KeywordDetected]++
			errorCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, fmt.Errorf("error scanning log file: %v", err)
	}

	return logEntries, keywordCounts, errorCount, nil
}

/******************************************************************************
* FUNCTION:        updateFileStats
*
* DESCRIPTION:     Update file_stats table with processing results
* INPUT:           fileID, status, startTime, errorCount, failureReason, keywordJSON
* RETURNS:         error
******************************************************************************/
func updateFileStats(fileID int64, status string, startTime time.Time, errorCount int, failureReason string, keywordJSON ...string) error {
	endTime := time.Now()
	processingTimeSec := endTime.Sub(startTime).Seconds()

	data := map[string]interface{}{
		"status":              status,
		"completed_at":        endTime,
		"processing_time_sec": processingTimeSec,
		"error_count":         errorCount,
		"process_start_time":  startTime,
	}

	if failureReason != "" {
		data["failure_reason"] = failureReason
	}

	if len(keywordJSON) > 0 && keywordJSON[0] != "" {
		data["keyword_stats"] = keywordJSON[0]
	}

	err := db.UpdateSingleRecord("file_stats", "file_id", fileID, data)
	if err != nil {
		fmt.Printf("Error updating file_stats: %v\n", err)
		return err
	}

	return nil
}
