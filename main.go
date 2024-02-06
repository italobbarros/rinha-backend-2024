// main.go
package main

import (
	"database/sql"
	"log"

	"github.com/italobbarros/rinha-backend-2024/internal/api"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres dbname=rinha password=postgres sslmode=disable"
	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	Api := api.NewApi(db)
	Api.Run()
}
