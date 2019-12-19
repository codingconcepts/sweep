package sweep

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type sweeper struct {
	db       *sql.DB
	c        Config
	Errs     chan error
	Affected chan int64
	Done     chan struct{}
}

// Config contains the configuration properties that will be used by
// the sweeper during calls to Sweep.
type Config struct {
	Interval          time.Duration
	IncrementInterval time.Duration
	Errors            chan error
	Affected          chan int64
	DeleteFunc        func() (string, []interface{})
}

// New returns a new sweeper object configured with the provide Config
// object.
func New(db *sql.DB, c Config) *sweeper {
	return &sweeper{
		db:   db,
		c:    c,
		Done: make(chan struct{}),
	}
}

// Sweep is a blocking function that clears down a database table using
// Config provided at initialisation. It will wait for a given Interval
// before starting a sweep process, clearing down Limit items at a time
// every IncrementInterval until the table is empty.
func (s *sweeper) Sweep() {
	interval := time.NewTicker(s.c.Interval).C
	for {
		select {
		case <-s.Done:
			return
		case <-interval:
			for range time.NewTicker(s.c.IncrementInterval).C {
				stmt, args := s.c.DeleteFunc()
				result, err := s.db.Exec(stmt, args...)
				if err != nil {
					sendError(s.c.Errors, err)
					break
				}

				affected, err := result.RowsAffected()
				if err != nil {
					sendError(s.c.Errors, err)
					break
				}
				sendAffected(s.c.Affected, affected)

				// Break out if we've cleared everything we need to for now.
				if affected == 0 {
					break
				}
			}
		}
	}
}

func sendError(c chan error, err error) {
	if c != nil {
		c <- err
	}
}

func sendAffected(c chan int64, affected int64) {
	if c != nil {
		c <- affected
	}
}
