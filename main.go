package main

import (
	"database/sql"
	"fmt"

	"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

type Product struct {
	Id         int
	Name       string
	Price      int
	ShortDesc  string
	LongDesc   string
	CategoryId int
	Image      string
}

var produts = []Product{}

func index(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM public.products")

	if err != nil {
		panic(err)
	}

	produts = []Product{}
	for res.Next() {
		var prd Product
		err = res.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
		if err != nil {
			panic(err)
		}
		produts = append(produts, prd)

	}

	tmpl, err := template.ParseFiles("html_files/index.html", "html_files/header.html")
	tmpl.ExecuteTemplate(page, "index", produts)
}

func main() {
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
	fmt.Println("Успешно подключено к базе данных")
}
