package main

import (
	"database/sql"
	"featuretester/pkg/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\tMAIN\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\tMAIN\t", log.Ldate|log.Ltime)

	dsn := "root:avito@tcp(localhost:3306)/featuretest?"
	dsn += "&charset=utf8"
	dsn += "&multiStatements=true"
	dsn += "&interpolateParams=true"
	dsn += "&parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		errLog.Printf("Couldn't start database driver: %s\n", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			errLog.Printf("Error closing database connection: %s\n", err)
		}
	}(db)
	db.SetMaxOpenConns(50)

	err = db.Ping()
	if err != nil {
		errLog.Printf("Couldn't connect to the database: %s\n", err)
	}

	featureHandler := handlers.NewFeaturesHandler(db)
	reportHandler := handlers.NewReportHandler(db)

	r := mux.NewRouter()
	r.HandleFunc("/api/create", featureHandler.AddFeature).Methods("POST")
	r.HandleFunc("/api/delete", featureHandler.DeleteFeature).Methods("DELETE")
	r.HandleFunc("/api/update_user_features", featureHandler.UpdateUserFeatures).Methods("POST")
	r.HandleFunc("/api/get_user_features", featureHandler.GetUserFeatures).Methods("GET")
	r.HandleFunc("/api/get_user_history", reportHandler.GetFeatureHistory).Methods("GET")

	r.PathPrefix("/reports/").Handler(
		http.StripPrefix("/reports/",
			http.FileServer(http.Dir("./static/reports"))))

	infoLog.Println("starting server at :8000")

	err = http.ListenAndServe("localhost:8000", r)
	if err != nil {
		errLog.Printf("listen and serve: %s\n", err)
	}
}
