package postgres

import (
	"database/sql"
	"log"
	"runtime"

	"github.com/lib/pq"
)

const (
	PG_UNIQUE_VIOLATION = pq.ErrorCode("23505")
)

type transactionBody func(tx *sql.Tx)

// runTransaction(db, body) wraps a transaction.
//
// The body function should panic on any error, using the error as the parameter to panic
func runTransaction(db *sql.DB, body transactionBody) (err error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("could not start a transaction. %v", err)
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	body(tx)

	err = tx.Commit()
	return
}

func isAlreadyExists(err error, table string) bool {
	if err, ok := err.(*pq.Error); ok {
		if err.Table == table && err.Code == PG_UNIQUE_VIOLATION {
			return true
		}
	}
	return false
}
