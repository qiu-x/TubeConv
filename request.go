package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	//"google.golang.org/api/googleapi/transport"
	//"google.golang.org/api/youtube/v3"
)

const developerKey = ""

func err_handle(err error) {
	log.Println(err)
	os.Exit(1)
}

func checklink_request(w http.ResponseWriter, body []byte) {
	checklink_req := struct {
		Request string
		Link    string
	}{}
	json.Unmarshal(body, &checklink_req)

	type checklink struct {
		Type string `json:"type"`
	}

	_, err := url.ParseRequestURI(checklink_req.Link)
	if err != nil {
		check_link := checklink{"title"}
		checklink_json, err := json.Marshal(check_link)
		if err != nil {
			err_handle(err)
		}
		w.Write(checklink_json)
		return
	}
	check_link := checklink{"link"}
	checklink_json, err := json.Marshal(check_link)
	if err != nil {
		err_handle(err)
	}
	w.Write(checklink_json)
}

func query_request(w http.ResponseWriter, body []byte) {
	query_req := struct {
		Request string
		Text    string
	}{}
	json.Unmarshal(body, &query_req)
	log.Println(query_req)

	text := strings.ReplaceAll(query_req.Text, " ", "+")
	parse, err := http.Get("https://www.youtube.com/results?search_query=" + text)
	if err != nil {
		err_handle(err)
	}
	defer parse.Body.Close()

	html, err := goquery.NewDocumentFromReader(parse.Body)
	if err != nil {
		err_handle(err)
	}

	index := strings.Index(html.Text(), "ytInitialData")
	ytInitialData := strings.Split(html.Text()[index:], "if (")[0]
	ytInitialData = strings.Split(ytInitialData, "ytInitialData =")[1]
	ytInitialData_byte := []byte(ytInitialData)[:len(ytInitialData) - 1]
	ytInitialData = string(ytInitialData_byte)

	var ytInitialDataJSON map[string]interface{}
	err = json.Unmarshal([]byte(ytInitialData), &ytInitialDataJSON)
	if err != nil {
		err_handle(err)
	}
	contents := ytInitialDataJSON["contents"].(map[string]interface{})["twoColumnSearchResultsRenderer"].
	(map[string]interface{})["primaryContents"].(map[string]interface{})["sectionListRenderer"].
	(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["itemSectionRenderer"].
	(map[string]interface{})["contents"].([]interface{})
	
	type video struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Link      string `json:"link"`
		Thumbnail string `json:"thumbnail"`
	}
	var vids []video

	for _, v := range contents {
		if v.(map[string]interface{})["videoRenderer"] == nil {
			continue
		}
		vid := video{v.(map[string]interface{})["videoRenderer"].(map[string]interface{})["title"].
		(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string), 
		v.(map[string]interface{})["videoRenderer"].(map[string]interface{})["longBylineText"].
		(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string),
		"https://www.youtube.com/watch?v=" + v.(map[string]interface{})["videoRenderer"].
		(map[string]interface{})["videoId"].(string), 
		v.(map[string]interface{})["videoRenderer"].(map[string]interface{})["thumbnail"].
		(map[string]interface{})["thumbnails"].([]interface{})[0].(map[string]interface{})["url"].(string)}

		vids = append(vids, vid)
	}
	videos := struct {
		Videos []video `json:"videos"`
	}{Videos: vids}
	videos_json, err := json.Marshal(videos)
	if err != nil {
		err_handle(err)
	}
	w.Write(videos_json)
}

//func query_request_old(w http.ResponseWriter, body []byte) {
//	query_req := struct {
//		Request string
//		Text    string
//	}{}
//	json.Unmarshal(body, &query_req)
//	log.Println(query_req)
//
//	client := &http.Client{
//		Transport: &transport.APIKey{Key: developerKey},
//	}
//	service, err := youtube.New(client)
//	if err != nil {
//		err_handle(err)
//	}
//
//	// Youtube API use
//	var look_by []string
//	look_by = append(look_by, "snippet")
//
//	call := service.Search.List(look_by).Q(query_req.Text).MaxResults(25)
//	result, err := call.Do()
//	if err != nil {
//		err_handle(err)
//	}
//
//	// Create and send respond
//	w.Header().Set("Content-Type", "application/json")
//	//var videos [5][4]string
//
//	type video struct {
//		Title     string `json:"title"`
//		Author    string `json:"author"`
//		Link      string `json:"link"`
//		Thumbnail string `json:"thumbnail"`
//	}
//	var vids []video
//
//	for _, item := range result.Items {
//		switch item.Id.Kind {
//		case "youtube#video":
//			vid := video{item.Snippet.Title, item.Snippet.ChannelTitle,
//				"https://youtube.com/watch?v=" + item.Id.VideoId, item.Snippet.Thumbnails.High.Url}
//			vids = append(vids, vid)
//		}
//	}
//	videos := struct {
//		Videos []video `json:"videos"`
//	}{Videos: vids}
//	videos_json, err := json.Marshal(videos)
//	if err != nil {
//		err_handle(err)
//	}
//	w.Write(videos_json)
//}
