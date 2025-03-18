/**************************************************************************
 * File       	   : dbHelperFunctions.go
 * DESCRIPTION     : This file contains functions that helps in different
*										 database operations
 * DATE            : 17-March-2025
 **************************************************************************/

package db

import (
	"LOGProcessor/shared/types"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

/******************************************************************************
* FUNCTION:        InitDbConnection
* DESCRIPTION:     Function to load env variables and assign to global variables
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func InitDbConnection() error {
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		types.CmnGlblCfg.DB_USER, types.CmnGlblCfg.DB_PASSWORD, types.CmnGlblCfg.DB_HOST, types.CmnGlblCfg.DB_PORT, types.CmnGlblCfg.DB_DATABASE)

	ctxH, err := sql.Open("postgres", connString)
	if err != nil {
		return err
	}

	if err := ctxH.Ping(); err != nil {
		return fmt.Errorf("error pinging db; err: %v", err)
	}

	types.Db.DbConn = ctxH
	return nil
}

/******************************************************************************
 * FUNCTION:        ExecQueryDB
 * DESCRIPTION:     This function will exec data from DB
 * INPUT:			query, whereEleList
 * RETURNS:    		outData, err
 ******************************************************************************/
func ExecQueryDB(query string, queryParams interface{}) (err error) {
	con := types.Db.DbConn

	_, err = con.Exec(query, queryParams)
	if err != nil {
		return err
	}
	return err
}

/******************************************************************************
 * FUNCTION:        AddMultipleRecordInDB
 * DESCRIPTION:     This function will insert multiple records in DB
 * INPUT:			tableName, data
 * RETURNS:    		err
 ******************************************************************************/
func AddMultipleRecordInDB(tableName string, data []map[string]interface{}) (err error) {

	if len(data) == 0 {
		return errors.New("empty data received")
	}

	con := types.Db.DbConn

	keys := make([]string, 0, len(data[0]))
	for key := range data[0] {
		keys = append(keys, key)
	}
	col := strings.Join(keys, ", ")

	var dataList []interface{}
	var placeholders []string

	for _, record := range data {
		var recordPlaceholders []string
		for _, key := range keys {
			recordPlaceholders = append(recordPlaceholders, fmt.Sprintf("$%d", len(dataList)+1))
			value := record[key]

			switch v := value.(type) {
			case []uint8:
				dataList = append(dataList, string(v))
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool:
				dataList = append(dataList, v)
			case time.Time:
				if v.IsZero() {
					dataList = append(dataList, nil)
				} else {
					dataList = append(dataList, v)
				}
			case nil:
				dataList = append(dataList, nil)
			default:
				dataList = append(dataList, fmt.Sprintf("%v", v))
			}
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(recordPlaceholders, ", ")))
	}

	query := fmt.Sprintf("INSERT INTO %v (%v) VALUES %v", tableName, col, strings.Join(placeholders, ", "))

	stmtIns, err := con.Prepare(query)
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(dataList...)
	if err != nil {
		return err
	}

	return nil
}

/******************************************************************************
 * FUNCTION:        UpdateDataInDB
 * DESCRIPTION:     This function will update data in DB
 * INPUT:
 * RETURNS:    		err, rows
 ******************************************************************************/
func UpdateDataInDB(query string, whereEleList []interface{}) (err error, rows int64) {

	rows = 0
	con := types.Db.DbConn

	stmtIns, err := con.Prepare(query)
	if err != nil {
		return err, rows
	}

	if stmtIns != nil {
		defer stmtIns.Close()
	}

	res, err := stmtIns.Exec(whereEleList...)
	if err != nil {
		return err, rows
	}

	rows, err = res.RowsAffected()
	return err, rows
}

/******************************************************************************
 * FUNCTION: GetDataFromDB
 * DESCRIPTION: This function will retrieve data from DB
 * INPUT: query, queryParams
 * RETURNS: results, err
 ******************************************************************************/
func GetDataFromDB(query string, queryParams []interface{}) (results []map[string]interface{}, err error) {
	con := types.Db.DbConn

	rows, err := con.Query(query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = make([]map[string]interface{}, 0)

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			switch v := val.(type) {
			case nil:
				rowMap[col] = nil
			case []byte:
				rowMap[col] = string(v)
			default:
				rowMap[col] = v
			}
		}

		results = append(results, rowMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
