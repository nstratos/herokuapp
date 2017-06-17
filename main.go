package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// App is the struct that holds details and services for the application.
type App struct {
	db *sql.DB
}

// Product is the struct that holds data for each Product.
type Product struct {
	Name    string
	Created time.Time
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	app := &App{db}

	http.Handle("/", http.HandlerFunc(app.serveHome))
	http.Handle("/add", http.HandlerFunc(app.addProduct))
	http.Handle("/products", http.HandlerFunc(app.showProducts))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (app *App) serveHome(w http.ResponseWriter, r *http.Request) {
	if _, err := app.db.Exec("CREATE TABLE IF NOT EXISTS products (name varchar(50), created timestamp)"); err != nil {
		http.Error(w, fmt.Sprintf("Error creating database table: %q", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}

func (app *App) showProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query("SELECT name, created FROM products")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading products: %q", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	products := make([]*Product, 0)
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.Name, &product.Created); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning products: %q", err), http.StatusInternalServerError)
			return
		}
		products = append(products, &product)
	}

	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding products: %q", err), http.StatusInternalServerError)
		return
	}
}

func (app *App) addProduct(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if _, err := app.db.Exec("INSERT INTO products VALUES ($1, now())", name); err != nil {
		http.Error(w, fmt.Sprintf("Error adding product: %q", err), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("success"))
}
