package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func IndexPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexPage)

	fmt.Println("starting server at :8000")
	err := http.ListenAndServe("localhost:8000", r)
	if err != nil {
		panic(err) // УБРАТЬ ПАНИКУ
	}
}
