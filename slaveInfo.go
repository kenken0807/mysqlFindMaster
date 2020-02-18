package main

import (
	"database/sql"
	"fmt"
)

func (slaveInfo *slaveStatusInfo) statusCheck() (bool, string) {
	if slaveInfo.sqlThread != "Yes" {
		return false, fmt.Sprintf("sqlThread [%s]", slaveInfo.sqlThread)
	}
	if slaveInfo.ioThread != "Yes" {
		return false, fmt.Sprintf("ioThread [%s]", slaveInfo.ioThread)
	}
	return true, ""

}

func columnIndex(cols []string, colName string) int {
	for idx := range cols {
		if cols[idx] == colName {
			return idx
		}
	}
	return -1
}

func columnValue(scanArgs []interface{}, cols []string, colName string) string {
	var columnIndex = columnIndex(cols, colName)
	if columnIndex == -1 {
		return ""
	}
	return string(*scanArgs[columnIndex].(*sql.RawBytes))
}
