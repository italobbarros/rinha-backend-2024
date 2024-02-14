// main.go
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/italobbarros/rinha-backend-2024/internal/api"
	_ "github.com/lib/pq"
)

func main() {
	connStr := fmt.Sprintf("host=%s port='5432' user='rinha' dbname='rinha' password='rinha' sslmode=disable", os.Getenv("DB_HOSTNAME"))
	var err error
	time.Sleep(10 * time.Second)
	db, err := sql.Open("postgres", connStr)
	for err != nil {
		time.Sleep(1 * time.Second)
		db, err = sql.Open("postgres", connStr)
	}
	err = db.Ping()
	if err != nil {
		for {
			time.Sleep(5 * time.Second)
			err := db.Ping()
			if err == nil {
				break
			}
			db, _ = sql.Open("postgres", connStr)
		}
	}

	db.SetMaxOpenConns(10)
	defer db.Close()
	Api := api.NewApi(db)
	Api.Run()
}
