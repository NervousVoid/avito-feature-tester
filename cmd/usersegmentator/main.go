package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"usersegmentator/config"
	"usersegmentator/pkg/errors"
	"usersegmentator/pkg/handlers"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

//	@title			Dynamic User Segmentation Service API
//	@version		1.0
//	@description	Avito Tech backend trainee assignment 2023

// @contact.name	Peter Androsov
// @contact.url	http://t.me/nervous_void
// @contact.email	androsov.p.v@gmail.com

func main() {
	infoLog := log.New(os.Stdout, "INFO\tMAIN\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\tMAIN\t", log.Ldate|log.Ltime)

	cfg, err := config.NewConfig()
	if err != nil {
		errLog.Printf("Error reading config: %s", err)
		return
	}

	dsn := fmt.Sprintf(
		"root:%s@tcp(%s:%s)/%s?",
		cfg.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Name,
	)
	dsn += "&charset=utf8"
	dsn += "&multiStatements=true"
	dsn += "&interpolateParams=true"
	dsn += "&parseTime=true"

	db, err := errors.DBConnectLoop(dsn, time.Duration(cfg.Timeout*1e9)) //nolint:gomnd // converting nanosecs to secs
	if err != nil {
		errLog.Printf("Couldn't start database driver: %s\n", err)
		return
	}

	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			errLog.Printf("Error closing database connection: %s\n", err)
		}
	}(db)
	db.SetMaxOpenConns(cfg.MaxConnections)

	segmentHandler := handlers.NewSegmentsHandler(db)
	reportHandler := handlers.NewHistoryHandler(db, cfg)

	r := mux.NewRouter()
	r.HandleFunc("/api/create_segment", segmentHandler.AddSegment).Methods("POST")
	r.HandleFunc("/api/delete_segment", segmentHandler.DeleteSegment).Methods("DELETE")
	r.HandleFunc("/api/update_user_segments", segmentHandler.UpdateUserSegments).Methods("POST")
	r.HandleFunc("/api/get_user_segments", segmentHandler.GetUserSegments).Methods("GET")
	r.HandleFunc("/api/get_user_history", reportHandler.GetUserHistory).Methods("GET")
	r.HandleFunc("/api/auto_assign_segment", segmentHandler.AutoAssignSegment).Methods("POST")

	r.PathPrefix("/reports/").Handler(
		http.StripPrefix("/reports/",
			http.FileServer(http.Dir("./"+cfg.StorageDir))))

	infoLog.Printf("starting server at %s:%s", cfg.HTTP.Host, cfg.HTTP.Port)

	err = http.ListenAndServe(cfg.HTTP.Host+":"+cfg.HTTP.Port, r)
	if err != nil {
		errLog.Printf("listen and serve: %s\n", err)
	}
}
