package main

import (
	"database/sql"
	"fmt"

	//"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	http.ListenAndServe(":8080", nil)
	fmt.Println("Успешно подключено к базе данных")
}
