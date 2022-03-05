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
	"flag"
)

var (
	port string
	ssl_full_path string
	ssl_priv_path string
	help bool
)

func init() {
	flag.StringVar(&ssl_full_path, "cf", "", "Path to the full chain SSL certificate")
	flag.StringVar(&ssl_priv_path, "cp", "", "Path to the private SSL certificate")
	flag.StringVar(&port, "p", "80", "Application Port")

	flag.StringVar(&ssl_full_path, "cert-full", "", "Path to the full chain SSL certificate")
	flag.StringVar(&ssl_priv_path, "cert-priv", "", "Path to the private SSL certificate")
	flag.StringVar(&port, "port", "80", "Application Port")
	flag.Parse()
}

type Request_type struct {
	Request string
}

var router = mux.NewRouter()

func simpleHandler(filepath string) func(http.ResponseWriter, *http.Request) {
	htmlRaw, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(htmlRaw)
	}
}

func fileServerFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "max-age=432000")
		next.ServeHTTP(w, r)
	})
}

type SafeMap struct{
	Mutex sync.Mutex
	Map  map[string]Download_data
}

type Download_data struct {
	Name   string
	File   io.ReadCloser
}

var Mapa SafeMap

func main() {
	Mapa.Map = make(map[string]Download_data)
	checkType := http.HandlerFunc(check_request_type)
	router.HandleFunc("/", simpleHandler("html/index.html"))
	router.HandleFunc("/req", checkType).Methods("POST")
	router.HandleFunc("/download/{id}", download_link_generator)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/res/", fileServerFilter(fs))
	log.Println("Listening on port: ", port)
	var err error
	if len(ssl_full_path) != 0 && len(ssl_priv_path) != 0 {
		err = http.ListenAndServeTLS(":"+port, ssl_full_path, ssl_priv_path, nil)
	} else {
		log.Println("Continuing without SSL")
		err = http.ListenAndServe(":"+port, nil)
	}
	if err != nil { log.Fatal(err) }

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
