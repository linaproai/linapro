// This file classifies MySQL driver errors for shared dialect callers.

package mysql

import (
	"errors"

	driver "github.com/go-sql-driver/mysql"
)

// Retryable write-conflict error codes reported by MySQL.
const (
	errorDeadlock        uint16 = 1213
	errorLockWaitTimeout uint16 = 1205
)

// IsRetryableWriteConflict classifies MySQL lock conflicts by stable server
// error numbers exposed by go-sql-driver/mysql.
func IsRetryableWriteConflict(err error) bool {
	var mysqlErr *driver.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}
	switch mysqlErr.Number {
	case errorDeadlock, errorLockWaitTimeout:
		return true
	default:
		return false
	}
}
