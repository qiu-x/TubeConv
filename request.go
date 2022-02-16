package main

import(
	"fmt"
	"os"
	"net/http"
	"net/url"
	"encoding/json"
	"flag"
	
	"google.golang.org/api/googleapi/transport"
    "google.golang.org/api/youtube/v3"
)

const developerKey = "AIzaSyAJ46IxN6B6u3BfWODc0AZFP3A1fdtBH-I"

func err_handle(err error){
	fmt.Println(err)
	os.Exit(1)
}

func checklink_request(w http.ResponseWriter, body []byte){
	checklink_req := struct {
		Request string
		Link string
	}{}
	json.Unmarshal(body, &checklink_req)

	type checklink struct{
		Type string `json:"type"`
	}

	_, err := url.ParseRequestURI(checklink_req.Link)
	if err != nil {
		check_link := checklink{"title"}
		checklink_json, err := json.Marshal(check_link)
		if err != nil {err_handle(err)}
		w.Write(checklink_json)
		return
	}
	check_link := checklink{"link"}
	checklink_json, err := json.Marshal(check_link)
	if err != nil {err_handle(err)}
	w.Write(checklink_json)
	
}

func query_request(w http.ResponseWriter, body []byte){
	query_req := struct {
		Request string
		Text string
	}{}
	json.Unmarshal(body, &query_req)
	
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}
	service, err := youtube.New(client)
	if err != nil {err_handle(err)}

	// Youtube API use
	flag.Parse()
	var look_by []string
	look_by = append(look_by, "snippet")

	query := flag.String("query", query_req.Text, "Search term")
	maxResults := flag.Int64("max-results", 3, "Max YouTube results")

	call := service.Search.List(look_by).Q(*query).MaxResults(*maxResults)
	result, err := call.Do()
	if err != nil {err_handle(err)}

	// Create and send respond
	w.Header().Set("Content-Type", "application/json")
	//var videos [5][4]string

	type video struct {
		Title string
		Author string
		Link string
		Thumbnail string
	}
	var vids []video

	for _, item := range result.Items {
		switch item.Id.Kind {
		case "youtube#video":
				vid := video{item.Snippet.Title, item.Snippet.ChannelTitle, 
				"https://youtube.com/watch?v=" + item.Id.VideoId, item.Snippet.Thumbnails.High.Url}
				vids = append(vids, vid)
		}
	}
	videos := struct { Videos []video `json:"videos"`}{Videos: vids}
	videos_json, err := json.Marshal(videos)
	if err != nil {err_handle(err)}
	w.Write(videos_json)
}