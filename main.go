// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
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
		log.Println(err)
		time.Sleep(1 * time.Second)
		db, err = sql.Open("postgres", connStr)
	}
	defer db.Close()
	Api := api.NewApi(db)
	Api.Run()
}
