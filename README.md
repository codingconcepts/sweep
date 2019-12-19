# sweep
Sweep away expired data from databases safely.

## Installation

```
$ go get -u -d github.com/codingconcepts/sweep
```

## Usage

``` go
// Sweep is a blocking function that will only return if you pass it
// a poison pill via its Done channel. If you'd like to report on how
// many items are deleted or how many errors occur during operation,
// this can be acheived with channels.
affected := make(chan int64, 10)
errors := make(chan error, 10)

// Inform sweep to wake up every hour and attempt to delete 1,000
// expired items from a database table. If more than zero items are
// deleted, it will continue to sweep every seconds, until there are
// no more expired items in the table. At which point, it sleeps for
// another hour.
s := New(db, Config{
	Affected:          make(chan int64, 10),
	Errors:            make(chan error, 10),
	Interval:          time.Hour * 1,
	IncrementInterval: time.Second * 10,
	DeleteFunc: func() (string, []interface{}) {
		return `DELETE FROM "your_table"
		WHERE "your_expiry_indicator" < $1
		LIMIT $2`, []interface{}{ time.Now().UTC(),	1000 }
	},
})

go s.Sweep()
```