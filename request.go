package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"sort"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"gopkg.in/validator.v2"
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

	link := strings.TrimSpace(checklink_req.Link)
	_, err := url.ParseRequestURI(link)
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
		Text    string `validate:"nonzero"`
	}{}
	json.Unmarshal(body, &query_req)
	if errs := validator.Validate(query_req); errs != nil {
		log.Println("Error:", errs)
		return
	}
	search_text := query_req.Text
	log.Println("Recived query request for:", search_text)

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
	ytInitialData_byte := []byte(ytInitialData)[:len(ytInitialData)-1]
	ytInitialData = string(ytInitialData_byte)

	var ytInitialDataJSON map[string]interface{}
	err = json.Unmarshal([]byte(ytInitialData), &ytInitialDataJSON)
	if err != nil {
		err_handle(err)
	}
	contents, check := ytInitialDataJSON["contents"].(map[string]interface{})["twoColumnSearchResultsRenderer"].(map[string]interface{})["primaryContents"].(map[string]interface{})["sectionListRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["itemSectionRenderer"].(map[string]interface{})["contents"].([]interface{})

	if !check {
		log.Println("Error: Getting data from yotube")
		return
	}

	type video struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Link      string `json:"link"`
		Thumbnail string `json:"thumbnail"`
	}
	var vids []video

	for _, v := range contents {
		videoRenderer, _ := v.(map[string]interface{})["videoRenderer"].(map[string]interface{})
		if videoRenderer == nil {
			continue
		}
		title, check := videoRenderer["title"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)
		if !check {
			log.Println("Error: title missing")
			return
		}
		author, check := videoRenderer["longBylineText"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)
		if !check {
			log.Println("Error: author missing")
			return
		}
		link, check := videoRenderer["videoId"].(string)
		if !check {
			log.Println("Error: link missing")
			return
		}
		link = "https://www.youtube.com/watch?v=" + link

		url, check := videoRenderer["thumbnail"].(map[string]interface{})["thumbnails"].([]interface{})[0].(map[string]interface{})["url"].(string)
		if !check {
			log.Println("Error: thumbnail missing")
			return
		}

		vid := video{title, author, link, url}

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
		Video_quality []string  `json:"video_quality"`
		Audio_quality []float64 `json:"audio_quality"`
	}

	var vid_info video_info
	vid_info.Audio_quality = append(vid_info.Audio_quality, 0)
	vid_info.Video_quality = append(vid_info.Video_quality, "none")

	var format_exist = make(map[string]int)
	for _, v := range formats {
		aserted_v, check := v.(map[string]interface{})
		if !check {
			log.Println("Error: Getting data from yt-dlp")
			return
		}

		storyboard, check := aserted_v["format_note"].(string)
		if !check {
			log.Println("Error: storyboard missing")
			return
		}
		if storyboard == "storyboard" {
			continue
		}
		_, exist := format_exist[aserted_v["format_note"].(string)]
		if exist {
			continue
		}
		audio_only, check := aserted_v["resolution"].(string)
		if !check {
			log.Println("Error: audio-only missing")
			return
		}
		if audio_only == "audio only" {
			abr, check := aserted_v["abr"].(float64)
			if !check {
				log.Println("Error: abr missing")
				return
			}
			vid_info.Audio_quality = append(vid_info.Audio_quality, abr)
		} else {
			format_exist[aserted_v["format_note"].(string)] = 0
			vid_info.Video_quality = append(vid_info.Video_quality, aserted_v["format_note"].(string))
		}
	}

	sort.Sort(sort.Reverse(sort.Float64Slice(vid_info.Audio_quality))) 
	vid_info_json, err := json.Marshal(vid_info)
	if err != nil {
		err_handle(err)
	}
	w.Write(vid_info_json)
}

func download_request(w http.ResponseWriter, body []byte) {
	download_req := struct {
		Request       string
		Link          string
		Video_quality string  `json:"video-quality"`
		Audio_quality float64 `json:"audio-quality"`
		Format        string
	}{}
	json.Unmarshal(body, &download_req)
	var download *exec.Cmd
	var ffmpeg *exec.Cmd
	var r io.ReadCloser
	var err error

	var audio_flag string
	var video_flag string

	switch {
	case download_req.Audio_quality != 0:
		audio_flag = "bestaudio[ext=m4a][abr<=" + strconv.FormatFloat(download_req.Audio_quality, 'f', 0, 64) + "]"
		break
	case download_req.Audio_quality != 0 && download_req.Format == "webm":
		audio_flag = "bestaudio[ext=webm][abr<=" + strconv.FormatFloat(download_req.Audio_quality, 'f', 0, 64) + "]"
		break
	case download_req.Audio_quality == 0:
		audio_flag = ""
		break
	}

	switch {
	case download_req.Video_quality != "none" && download_req.Format != "mp3" && download_req.Format != "ogg":
		video_flag = "bestvideo[ext=" + download_req.Format +"][height<="+download_req.Video_quality+"]"
		break
	case download_req.Video_quality == "none":
		video_flag = ""
		break
	}

	if len(audio_flag) <= 1 || len(video_flag) <= 1 {
		download = exec.Command("yt-dlp", "-f", video_flag+audio_flag,
		"-o", "-", download_req.Link)
	}else{
		download = exec.Command("yt-dlp", "-f", video_flag+"+"+audio_flag,
		"-o", "-", download_req.Link)
		log.Println("yt-dlp -f" + video_flag+audio_flag + 
		"-o -" + download_req.Link)
	}

	switch {
	case download_req.Format == "ogg":
		ffmpeg = exec.Command("ffmpeg", "-i", "-", "-f", "ogg", "-")
		ffmpeg.Stdin, _ = download.StdoutPipe()
		r, err = ffmpeg.StdoutPipe()
		break
	case download_req.Format == "mp3":
		ffmpeg = exec.Command("ffmpeg", "-i", "-", "-f", "ogg", "-")
		ffmpeg.Stdin, _ = download.StdoutPipe()
		r, err = ffmpeg.StdoutPipe()
		break
	default:
		r, err = download.StdoutPipe()
		break
	}

	if err != nil {
		err_handle(err)
	}

	download.Start()
	go func() {
		download.Process.Wait()
		download.Process.Kill()
	}()
	if ffmpeg != nil {
		ffmpeg.Start()
		go func() {
			ffmpeg.Process.Wait()
			ffmpeg.Process.Kill()
		}()
	}

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 25)
	for i := range id {
		id[i] = charset[seededRand.Intn(len(charset))]
	}

	file_name, err := exec.Command("yt-dlp", "--get-filename", download_req.Link).Output()
	if err != nil {
		err_handle(err)
	}
	filename_arr := strings.Split(string(file_name), ".")
	filename := strings.Join(filename_arr[:len(filename_arr)-1], "") + "." + download_req.Format

	Mapa.Mutex.Lock()
	if ffmpeg != nil {
		Mapa.Map[string(id)] = Download_data{filename, r, ffmpeg}
	}else{
		Mapa.Map[string(id)] = Download_data{filename, r, download}
	}
	Mapa.Mutex.Unlock()
	down_link := struct {
		File string `json:"file"`
	}{
		"/download/" + string(id),
	}
	down_link_json, err := json.Marshal(down_link)
	if err != nil {
		err_handle(err)
	}
	w.Write(down_link_json)
}

func download_link_generator(w http.ResponseWriter, req *http.Request) {
	Mapa.Mutex.Lock()
	vars := mux.Vars(req)
	r, exist := Mapa.Map[vars["id"]]
	if !exist {
		http.NotFound(w, req)
		return
	}
	delete(Mapa.Map, vars["id"])
	Mapa.Mutex.Unlock()
	w.Header().Set("content-disposition", "attachment; filename="+r.Name)
	buffer := make([]byte, 1024)
	io.CopyBuffer(w, r.File, buffer)
	r.File.Close()
	r.Command.Process.Wait()
	r.Command.Process.Kill()	
}
