package main

import(
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type Request_type struct {
	Request string
}

func main(){
	checkType := http.HandlerFunc(check_request_type)
	http.Handle("/", checkType)
	http.ListenAndServe(":8080", nil)
}

func check_request_type(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, r.URL.Path[1:])
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {err_handle(err)}
	var req_type Request_type
	json.Unmarshal(body, &req_type)

	if req_type.Request == "check-link" {
		checklink_request(w, body)
	}else if req_type.Request == "query" {
		query_request(w, body)
	}
}