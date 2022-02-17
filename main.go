package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Request_type struct {
	Request string
}

var router = mux.NewRouter()

func simpleHandler(filepath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		htmlRaw, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
			http.NotFound(w, r)
			return
		}
		w.Write(htmlRaw)
	}
}

func fileServerFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	checkType := http.HandlerFunc(check_request_type)
	router.HandleFunc("/", simpleHandler("html/index.html"))
	router.HandleFunc("/req", checkType).Methods("POST")
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/res/", fileServerFilter(fs))
	http.ListenAndServe(":8080", nil)
}

func check_request_type(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err_handle(err)
	}
	var req_type Request_type
	json.Unmarshal(body, &req_type)

	if req_type.Request == "check-link" {
		checklink_request(w, body)
	} else if req_type.Request == "query" {
		query_request(w, body)
	}
}
