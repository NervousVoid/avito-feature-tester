package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"usersegmentator/pkg/handlers"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const maxDBConnections = 50

func main() {
	infoLog := log.New(os.Stdout, "INFO\tMAIN\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\tMAIN\t", log.Ldate|log.Ltime)

	dsn := "root:avito@tcp(localhost:3306)/usersegmentator?"
	dsn += "&charset=utf8"
	dsn += "&multiStatements=true"
	dsn += "&interpolateParams=true"
	dsn += "&parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		errLog.Printf("Couldn't start database driver: %s\n", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			errLog.Printf("Error closing database connection: %s\n", err)
		}
	}(db)
	db.SetMaxOpenConns(maxDBConnections)

	err = db.Ping()
	if err != nil {
		errLog.Printf("Couldn't connect to the database: %s\n", err)
	}

	segmentHandler := handlers.NewSegmentsHandler(db)
	reportHandler := handlers.NewHistoryHandler(db)

	r := mux.NewRouter()
	r.HandleFunc("/api/create_segment", segmentHandler.AddSegment).Methods("POST")
	r.HandleFunc("/api/delete_segment", segmentHandler.DeleteSegment).Methods("DELETE")
	r.HandleFunc("/api/update_user_segments", segmentHandler.UpdateUserSegments).Methods("POST")
	r.HandleFunc("/api/get_user_segments", segmentHandler.GetUserSegments).Methods("GET")
	r.HandleFunc("/api/get_user_history", reportHandler.GetSegmentHistory).Methods("GET")
	r.HandleFunc("/api/auto_assign_segments", segmentHandler.AutoAssignSegment).Methods("POST")

	r.PathPrefix("/reports/").Handler(
		http.StripPrefix("/reports/",
			http.FileServer(http.Dir("./static/reports"))))

	infoLog.Println("starting server at :8000")

	err = http.ListenAndServe("localhost:8000", r)
	if err != nil {
		errLog.Printf("listen and serve: %s\n", err)
	}
}
