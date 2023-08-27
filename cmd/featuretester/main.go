package main

import (
	"database/sql"
	"featuretester/pkg/feature"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:avito@tcp(localhost:3306)/featuretest?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Couldn't start database driver: %v\n", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("Error closing database connection\n")
		}
	}(db)
	db.SetMaxOpenConns(50)

	err = db.Ping()
	if err != nil {
		fmt.Printf("Couldn't connect to the database: %v\n", err)
	}

	featureHandler := feature.Handler{DB: db}

	r := mux.NewRouter()

	r.HandleFunc("/api/create", featureHandler.AddFeature).Methods("POST")
	//r.HandleFunc("/api/delete", feature.DeleteFeature).Methods("DELETE")

	fmt.Println("starting server at :8000")
	err = http.ListenAndServe("localhost:8000", r)
	if err != nil {
		panic(err) // УБРАТЬ ПАНИКУ
	}
}
