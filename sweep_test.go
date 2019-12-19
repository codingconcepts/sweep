package sweep

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	db   *sql.DB
	mock sqlmock.Sqlmock
)

func TestMain(m *testing.M) {
	var err error
	if db, mock, err = sqlmock.New(); err != nil {
		log.Fatalf("error creating database mock: %v", err)
	}

	os.Exit(m.Run())
}

func TestSweep(t *testing.T) {
	now := time.Now().UTC()

	affectedC := make(chan int64, 10)
	s := New(db, Config{
		Affected:          affectedC,
		Interval:          time.Millisecond * 200,
		IncrementInterval: time.Millisecond * 100,
		DeleteFunc: func() (string, []interface{}) {
			return `DELETE FROM "reservation" WHERE "expiry" < $1 LIMIT $2`, []interface{}{
				now,
				1000,
			}
		},
	})

	mock.ExpectExec("DELETE FROM").WithArgs(now, 1000).WillReturnResult(sqlmock.NewResult(1, 1000))
	mock.ExpectExec("DELETE FROM").WithArgs(now, 1000).WillReturnResult(sqlmock.NewResult(1, 999))
	mock.ExpectExec("DELETE FROM").WithArgs(now, 1000).WillReturnResult(sqlmock.NewResult(1, 0))

	go s.Sweep()

	equals(t, int64(1000), <-affectedC)
	equals(t, int64(999), <-affectedC)
	equals(t, int64(0), <-affectedC)

	s.Done <- struct{}{}
	errorNil(t, mock.ExpectationsWereMet())
}

func TestSweepErrorExecuting(t *testing.T) {
	now := time.Now().UTC()

	errorC := make(chan error, 10)
	s := New(db, Config{
		Errors:            errorC,
		Interval:          time.Millisecond * 200,
		IncrementInterval: time.Millisecond * 100,
		DeleteFunc: func() (string, []interface{}) {
			return `DELETE FROM "reservation" WHERE "expiry" < $1 LIMIT $2`, []interface{}{
				now,
				1000,
			}
		},
	})

	err := errors.New("oh noes!")
	mock.ExpectExec("DELETE FROM").WithArgs(now, 1000).WillReturnError(err)

	go s.Sweep()

	equals(t, err, <-errorC)

	s.Done <- struct{}{}
	errorNil(t, mock.ExpectationsWereMet())
}

func TestSweepErrorRowsAffected(t *testing.T) {
	now := time.Now().UTC()

	errorC := make(chan error, 10)
	s := New(db, Config{
		Errors:            errorC,
		Interval:          time.Millisecond * 200,
		IncrementInterval: time.Millisecond * 100,
		DeleteFunc: func() (string, []interface{}) {
			return `DELETE FROM "reservation" WHERE "expiry" < $1 LIMIT $2`, []interface{}{
				now,
				1000,
			}
		},
	})

	err := errors.New("oh noes!")
	mock.ExpectExec("DELETE FROM").WithArgs(now, 1000).WillReturnResult(sqlmock.NewErrorResult(err))

	go s.Sweep()

	equals(t, err, <-errorC)

	s.Done <- struct{}{}
	errorNil(t, mock.ExpectationsWereMet())
}

func equals(tb testing.TB, exp, act interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		tb.Fatalf("\nexp:\t%[1]v (%[1]T)\ngot:\t%[2]v (%[2]T)", exp, act)
	}
}

func errorNil(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}
