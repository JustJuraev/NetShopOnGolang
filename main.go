package main

import (
	"database/sql"
	"fmt"
	"time"

	//"encoding/gob"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

type app struct {
}

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
	UserId   int
}

type ProductProperty struct {
	Id            int
	ProductId     int
	PropertyName  string
	PropertyValue string
	CategoryId    int
}

type User struct {
	Id       int
	Name     string
	Password string
	Number   string
}

type OrderItem struct {
	Id           int
	ProductId    int
	ProductName  string
	ProductCount int
	OrderId      int
}

var produts = []Product{}
var categories = []Category{}
var productsbycat = []Product{}
var basket = []BasketProduct{}
var productProperties = []ProductProperty{}
var productPropertiesbycat = []ProductProperty{}
var ppbyfilter = []ProductProperty{}

var str string
var flag bool

var store = sessions.NewCookieStore([]byte("basket-secret"))
var cache = map[string]User{}

func containsProduct(product []Product, prd Product) bool {
	for _, v := range product {
		if v.Id == prd.Id {
			return true
		}
	}

	return false
}

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

func filter(page http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	s := r.Form["chx"]

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	ppbyfilter = []ProductProperty{}
	for _, v := range s {
		row, err := db.Query("SELECT * FROM public.productproperties WHERE propertyvalue = $1", v)

		for row.Next() {
			var p ProductProperty
			err = row.Scan(&p.Id, &p.ProductId, &p.PropertyName, &p.PropertyValue, &p.CategoryId)
			if err != nil {
				panic(err)
			}
			ppbyfilter = append(ppbyfilter, p)
		}
	}

	var id int
	productsbycat = []Product{}
	for _, v := range ppbyfilter {
		//productsbycat = []Product{}
		row := db.QueryRow("SELECT * FROM public.products WHERE id = $1", v.ProductId)

		prd := Product{}
		err = row.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
		if err != nil {
			panic(err)
		}
		if containsProduct(productsbycat, prd) == false {
			productsbycat = append(productsbycat, prd)
			id = prd.CategoryId
		}
	}

	row2, err := db.Query("SELECT * FROM public.productproperties WHERE categoryid = $1", id)

	productPropertiesbycat = []ProductProperty{}
	for row2.Next() {
		var pr ProductProperty
		err = row2.Scan(&pr.Id, &pr.ProductId, &pr.PropertyName, &pr.PropertyValue, &pr.CategoryId)
		if err != nil {
			panic(err)
		}

		productPropertiesbycat = append(productPropertiesbycat, pr)

	}

	data := struct {
		Array1 []ProductProperty
		Array2 []Product
	}{
		Array2: productsbycat,
		Array1: productPropertiesbycat,
	}

	tmpl, err := template.ParseFiles("html_files/category.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "category", data)

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
		tmpl, err := template.ParseFiles("html_files/basket.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		basket = []BasketProduct{}
		tmpl.ExecuteTemplate(page, "basket", basket)
		return
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
	//fmt.Println(session.Values["basket"])
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

	row2, err := db.Query("SELECT * FROM public.productproperties WHERE categoryid = $1", id)

	productPropertiesbycat = []ProductProperty{}
	for row2.Next() {
		var pr ProductProperty
		err = row2.Scan(&pr.Id, &pr.ProductId, &pr.PropertyName, &pr.PropertyValue, &pr.CategoryId)
		if err != nil {
			panic(err)
		}

		productPropertiesbycat = append(productPropertiesbycat, pr)

	}

	data := struct {
		Array1 []ProductProperty
		Array2 []Product
	}{
		Array2: productsbycat,
		Array1: productPropertiesbycat,
	}

	tmpl, err := template.ParseFiles("html_files/category.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "category", data)
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
		panic(err)
	}

	res2, err := db.Query("SELECT * FROM public.productproperties WHERE productid = $1", id)

	if err != nil {
		panic(err)
	}

	productProperties = []ProductProperty{}
	for res2.Next() {
		var pr ProductProperty
		err = res2.Scan(&pr.Id, &pr.ProductId, &pr.PropertyName, &pr.PropertyValue, &pr.CategoryId)

		if err != nil {
			panic(err)
		}

		productProperties = append(productProperties, pr)
	}

	data := struct {
		Product Product
		Array   []ProductProperty
	}{
		Array:   productProperties,
		Product: prd,
	}

	tmpl, err := template.ParseFiles("html_files/product.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "product", data)
}

func LogOut(page http.ResponseWriter, r *http.Request) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1}
		http.SetCookie(page, &c)
	}
	http.Redirect(page, r, "/", http.StatusSeeOther)
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

func readCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		return
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return value, err
	}
	str := cookie.Value
	value, _ = url.QueryUnescape(str)
	return value, err
}

func LoginPage(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/login.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "login", nil)
}

func LoginCheck(page http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	//	fmt.Println(login)
	//	fmt.Println(password)

	if login == "" || password == "" {
		tmpl, err := template.ParseFiles("html_files/login.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "login", "Имя или пароль не может быть пустым")
		return
	}
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	defer db.Close()
	res := db.QueryRow("SELECT * FROM public.users WHERE name = $1 AND password = $2", login, hashedPass)
	user := User{}
	err = res.Scan(&user.Id, &user.Name, &user.Password)
	if err != nil {
		tmpl, err := template.ParseFiles("html_files/login.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "login", "неверный логин или пароль")
		return

	}

	token := login
	hashToken := md5.Sum([]byte(token))
	hashedToken := hex.EncodeToString(hashToken[:])
	cache[hashedToken] = user
	livingTime := 120 * time.Hour
	expiration := time.Now().Add(livingTime)

	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(hashedToken), Expires: expiration}
	http.SetCookie(page, &cookie)
	flag = true
	http.Redirect(page, r, "/", http.StatusSeeOther)
}

func RegisterPage(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/register.html", "html_files/header.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "register", nil)
}

func RegisterCheck(page http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	confirmpassword := r.FormValue("confirmpassword")

	if login == "" || password == "" || confirmpassword == "" {
		tmpl, err := template.ParseFiles("html_files/register.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "register", "все поля должны быть заполнены!")
		return
	}

	if password != confirmpassword {
		tmpl, err := template.ParseFiles("html_files/register.html", "html_files/header.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "register", "пароли не совпадают")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO public.users (name, password) VALUES ($1, $2)", login, hashedPass)

	http.Redirect(page, r, "/login", http.StatusSeeOther)
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

	cookie, err := readCookie("token", r)

	if err != nil {
		panic(err)
	}
	us := cache[cookie]

	_, err = db.Exec("INSERT INTO public.orders (address, number, cartnum, delivery, time, userid) VALUES ($1, $2, $3, $4, $5, $6)", address, number, cartnum, delivery, today, us.Id)
	row := db.QueryRow("SELECT * FROM public.orders WHERE time = $1", today)
	//fmt.Println(id)
	ord := Order{}
	err = row.Scan(&ord.Id, &ord.Address, &ord.Delivery, &ord.Number, &ord.CartNum, &ord.Time, &ord.UserId)
	if err != nil {
		panic(err)
	}

	for _, v := range basket {
		_, err = db.Exec("INSERT INTO public.orderitems (productid, productname, productcount, orderid) VALUES ($1, $2, $3, $4)", v.Id, v.Name, v.Count, ord.Id)
	}

	session, _ := store.Get(r, "sesssion")
	session.Options.MaxAge = -1
	session.Save(r, page)
	http.Redirect(page, r, "/", http.StatusSeeOther)

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
	router.HandleFunc("/Filter", filter)
	router.HandleFunc("/login", LoginPage)
	router.HandleFunc("/login_check", LoginCheck)
	router.HandleFunc("/logout", LogOut)
	router.HandleFunc("/register", RegisterPage)
	router.HandleFunc("/register_check", RegisterCheck)
	router.HandleFunc("/product/{id:[0-9]+}", GetProduct).Methods("GET")
	router.HandleFunc("/category/{id:[0-9]+}", GetByCategoty).Methods("GET")
	router.HandleFunc("/AddToBasket/{id:[0-9]+}", AddToBasket).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)

}
