package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"strconv"
	"io"
	"math/rand"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

func err_handle(err error) {
	log.Println(err)
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

func videoinfo_request(w http.ResponseWriter, body []byte) {
	videoinfo_req := struct {
		Request string
		Link    string
	}{}
	json.Unmarshal(body, &videoinfo_req)

	output, err := exec.Command("yt-dlp", "-J", videoinfo_req.Link).Output()
	if err != nil {
		err_handle(err)
	}
	var output_json map[string]interface{}
	json.Unmarshal(output, &output_json)
	formats := output_json["formats"].([]interface{})

	type video_info struct {
		Video_quality    []string `json:"video_quality"`
		Audio_quality    []float64 `json:"audio_quality"`
	}

	var vid_info video_info
	var format_exist = make(map[string]int)
	for _, v := range formats {
		if v.(map[string]interface{})["format_note"].(string) == "storyboard" {
			continue
		}
		_, exist := format_exist[v.(map[string]interface{})["format_note"].(string)]
		if exist {
			continue
		}

		if v.(map[string]interface{})["resolution"].(string) == "audio only" {
			vid_info.Audio_quality = append(vid_info.Audio_quality, v.
			(map[string]interface{})["abr"].(float64))
		}else {
			format_exist[v.(map[string]interface{})["format_note"].(string)] = 0
			vid_info.Video_quality = append(vid_info.Video_quality, v.
			(map[string]interface{})["format_note"].(string))
		}
	}
	vid_info.Audio_quality = append(vid_info.Audio_quality, 0)
	vid_info.Video_quality = append(vid_info.Video_quality, "none")

	vid_info_json, err := json.Marshal(vid_info)
	if err != nil {
		err_handle(err)
	}
	w.Write(vid_info_json)
}

func download_request(w http.ResponseWriter, body []byte) {
	download_req := struct {
		Request 	   string
		Link    	   string
		Video_quality  string `json:"video-quality"`
		Audio_quality  float64 `json:"audio-quality"`
		Format		   string
	}{}
	json.Unmarshal(body, &download_req)
	var download *exec.Cmd
	var ffmpeg *exec.Cmd
	var r io.ReadCloser
	var err error

	if download_req.Audio_quality != 0 && download_req.Video_quality != "none" {
		download = exec.Command("yt-dlp", "-f", "bestvideo[ext=" + download_req.Format + "][height<=" + 
		download_req.Video_quality + "]+bestaudio[ext=m4a][abr<=" + 
		strconv.FormatFloat(download_req.Audio_quality, 'f', 0, 64) + "]",
		"-o", "-", download_req.Link)
		r, err = download.StdoutPipe()
	}

	if download_req.Audio_quality != 0 && download_req.Video_quality == "none" {
		download = exec.Command("yt-dlp", "-f", "bestaudio[ext=m4a][abr<=" + 
		strconv.FormatFloat(download_req.Audio_quality, 'f', 0, 64) + "]", "-o", "-", download_req.Link)
		if download_req.Format != "mp4" {
			ffmpeg = exec.Command("ffmpeg", "-i", "-", "-f", "mp3", "-")
			ffmpeg.Stdin, _ = download.StdoutPipe()
			r, err = ffmpeg.StdoutPipe()
		}else{
			r, err = download.StdoutPipe()
		}
	}

	if download_req.Audio_quality == 0 && download_req.Video_quality != "none" {
		download = exec.Command("yt-dlp", "-f", "bestvideo[ext=" + download_req.Format + "][height<=" + 
		download_req.Video_quality + "]", "-o", "-", download_req.Link)
		r, err = download.StdoutPipe()
	}
	if err != nil {
		err_handle(err)
	}
	download.Start()
	if ffmpeg != nil {
		ffmpeg.Start()
	}

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))	  
	const charset = "abcdefghijklmnopqrstuvwxyz" +
  	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 25)
  	for i := range id {
    	id[i] = charset[seededRand.Intn(len(charset))]
  	}
	log.Println(string(id))
	Mapa.Mutex.Lock()
	Mapa.Map[string(id)] = Download_data{download_req.Format, r}
	Mapa.Mutex.Unlock()
}

func download_link_generator(w http.ResponseWriter, r *http.Request) {
	Mapa.Mutex.Lock()
	defer Mapa.Mutex.Unlock()
	vars := mux.Vars(r)
	_, exist := Mapa.Map[vars["id"]]
	if !exist {
		http.NotFound(w,r)
		return
	}
	w.Header().Set("content-disposition", "attachment; filename=example."+ Mapa.Map[vars["id"]].Format)
	buffer := make([]byte, 1024)
	io.CopyBuffer(w, Mapa.Map[vars["id"]].File, buffer)
	delete(Mapa.Map, vars["id"])
}