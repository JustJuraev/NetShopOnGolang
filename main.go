package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	//"encoding/gob"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

type Product struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	ShortDesc  string `json:"shortdesc"`
	LongDesc   string `json:"longdesc"`
	CategoryId int    `json:"categoryid"`
	Image      string `json:"image"`
}
type Category struct {
	Id    int
	Name  string
	Image string
}

type BasketProduct struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	CategoryId int    `json:"categoryid"`
	Image      string `json:"image"`
	Count      int    `json:"count"`
}

type Order struct {
	Id       int
	Address  string
	Delivery bool
	Number   string
	CartNum  string
	Time     time.Time
}

var produts = []Product{}
var categories = []Category{}
var productsbycat = []Product{}
var basket = []BasketProduct{}

var str string

var store = sessions.NewCookieStore([]byte("basket-secret"))

func contains(product []BasketProduct, prd Product) bool {
	for _, v := range product {
		if v.Id == prd.Id {
			return true
		}
	}

	return false
}

func GetIndex(product []BasketProduct, prd Product) int {

	var index int

	for idx, v := range product {
		if v.Id == prd.Id {
			index = idx
		}
	}

	return index
}

func GetProductFromBasket(product []BasketProduct, prd Product) BasketProduct {

	var pr BasketProduct

	for _, v := range product {
		if v.Id == prd.Id {
			pr = v
		}
	}

	return pr
}

func AddToBasket(page http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	//fmt.Println(id)

	row := db.QueryRow("SELECT * FROM public.products WHERE id = $1", id)
	//fmt.Println(id)
	prd := Product{}
	err = row.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
	if err != nil {
		http.Error(page, http.StatusText(404), http.StatusNotFound)
		fmt.Println(err)
	} else {
		if contains(basket, prd) == false {
			basketprd := BasketProduct{
				Id:         prd.Id,
				Name:       prd.Name,
				Price:      prd.Price,
				CategoryId: prd.CategoryId,
				Image:      prd.Image,
				Count:      1,
			}
			basket = append(basket, basketprd)
			//fmt.Println(basket)
		} else {
			var s = GetProductFromBasket(basket, prd)
			s.Count++
			var idx = GetIndex(basket, prd)
			basket[idx] = basket[len(basket)-1]
			basket[len(basket)-1] = BasketProduct{}
			basket = basket[:len(basket)-1]
			basket = append(basket, s)
		}

	}

	jsonBytes, err := json.Marshal(&basket)
	session, _ := store.Get(r, "sesssion")
	session.Values["basket"] = string(jsonBytes)
	session.Save(r, page)

	http.Redirect(page, r, "/basket", http.StatusSeeOther)

}

func ShowBasket(page http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "sesssion")
	untyped, ok := session.Values["basket"]

	if !ok {
		panic(ok)
	}

	basket, ok := untyped.(string)
	//	fmt.Println(basket)

	basketP := []BasketProduct{}
	err := json.Unmarshal([]byte(basket), &basketP)

	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles("html_files/basket.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "basket", basketP)
}

func GetByCategoty(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err := db.Query("SELECT * FROM public.products WHERE categoryid = $1", id)

	productsbycat = []Product{}
	for row.Next() {
		var prd Product
		err = row.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
		if err != nil {
			panic(err)
		}

		productsbycat = append(productsbycat, prd)

	}

	tmpl, err := template.ParseFiles("html_files/category.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "category", productsbycat)
}

func GetProduct(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row := db.QueryRow("SELECT * FROM public.products WHERE id = $1", id)
	//fmt.Println(id)
	prd := Product{}
	err = row.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
	if err != nil {
		http.Error(page, http.StatusText(404), http.StatusNotFound)
		fmt.Println(err)
	} else {
		tmpl, err := template.ParseFiles("html_files/product.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "product", prd)
	}
}

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

	defer res.Close()

	res2, err := db.Query("SELECT * FROM public.categories")

	if err != nil {
		panic(err)
	}

	categories = []Category{}
	for res2.Next() {
		var cat Category
		err = res2.Scan(&cat.Id, &cat.Name, &cat.Image)
		if err != nil {
			panic(err)
		}
		categories = append(categories, cat)
	}

	data := struct {
		Array1 []Category
		Array2 []Product
	}{
		Array2: produts,
		Array1: categories,
	}

	tmpl, err := template.ParseFiles("html_files/index.html", "html_files/header.html")
	tmpl.ExecuteTemplate(page, "index", data)
}

func saveOrder(page http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	number := r.FormValue("number")
	cartnum := r.FormValue("cartnum")
	delivery := r.FormValue("delivery")
	today := time.Now()

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO public.orders (address, number, cartnum, delivery, time) VALUES ($1, $2, $3, $4, $5)", address, number, cartnum, delivery, today)

	http.Redirect(page, r, "/", http.StatusSeeOther)

	// defer insert.Scan()

}

func main() {
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router := mux.NewRouter()

	router.HandleFunc("/", index)

	router.HandleFunc("/basket", ShowBasket)
	router.HandleFunc("/saveOrder", saveOrder)
	router.HandleFunc("/product/{id:[0-9]+}", GetProduct).Methods("GET")
	router.HandleFunc("/category/{id:[0-9]+}", GetByCategoty).Methods("GET")
	router.HandleFunc("/AddToBasket/{id:[0-9]+}", AddToBasket).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)

}
