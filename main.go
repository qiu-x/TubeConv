package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
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

type SafeMap struct{
	Mutex sync.Mutex
	Map  map[string]Download_data
}

type Download_data struct {
	Format string
	File   io.Reader
}

var Mapa SafeMap

func main() {
	Mapa.Map = make(map[string]Download_data)
	checkType := http.HandlerFunc(check_request_type)
	router.HandleFunc("/res/icons/favicon-32x32.png", simpleHandler("res/icons/favicon-32x32.png"))
	router.HandleFunc("/res/icons/favicon-16x16.png", simpleHandler("res/icons/favicon-16x16.png"))
	router.HandleFunc("/", simpleHandler("html/index.html"))
	router.HandleFunc("/req", checkType).Methods("POST")
	router.HandleFunc("/download/{id}", download_link_generator)
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
	}else if req_type.Request == "query" {
		query_request(w, body)
	}else if req_type.Request == "video-info" {
		videoinfo_request(w, body)
	}else if req_type.Request == "download" {
		download_request(w, body)
	}
}
