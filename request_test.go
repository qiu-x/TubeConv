package main

import (
	"testing"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"net/http"
	"io"
	"os/exec"
	"io/ioutil"
	"bytes"

	"github.com/gorilla/mux"
)

func TestChecklink(t *testing.T) {
	type req_template struct{
		Request string `json:"request"`
		Link string `json:"link"`
	}
	type respond_template struct {
		Type string `json:"type"`
	}

	tests := make(map[req_template]respond_template)

	req := req_template{"check-link", "https://www.youtube.com/watch?v=lTAXnaWMR3s"}
	tests[req] = respond_template{"link"}
	req = req_template{"check-link", "https://youtu.be/lTAXnaWMR3s"}
	tests[req] = respond_template{"link"}
	req = req_template{"check-link", "    https://youtu.be/lTAXnaWMR3s    "}
	tests[req] = respond_template{"link"}
	req = req_template{"check-link", "https://youtu.be/lTAXnaWMR3s?t=6"}
	tests[req] = respond_template{"link"}
	req = req_template{"check-link", "https://www.youtube.com/shorts/r3VI_rAxivM"}
	tests[req] = respond_template{"link"}

	req = req_template{"check-link", "good grief"}
	tests[req] = respond_template{"title"}

	for req, respond := range tests {
		r := httptest.NewRecorder()
		req_json, err := json.Marshal(req)
		if err != nil {
			err_handle(err)
		}
		checklink_request(r, req_json)
		respond_json, err := json.Marshal(respond)
		if err != nil {
			err_handle(err)
		}
		if string(respond_json) != r.Body.String() {
			t.Errorf("Want: %s Gets: %s", string(respond_json), r.Body.String())
			t.Fail()
		}
	}
}

func TestQuery(t *testing.T) {
	type req_template struct{
		Request string
		Text    string
	}
	type respond_template struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Link      string `json:"link"`
		Thumbnail string `json:"thumbnail"`
	}

	tests := make(map[req_template]respond_template)
	
	req := req_template{"query", "https://www.youtube.com/watch?v=hOrbfQpdLKo"}
	tests[req] = respond_template{"Bastille - Good Grief (Clean Version - Official Music Video)", 
	"BASTILLEvideos", "https://www.youtube.com/watch?v=hOrbfQpdLKo", "https://i.ytimg.com/vi/hOrbfQpdLKo/hq720.jpg?sqp=-oaymwEjCOgCEMoBSFryq4qpAxUIARUAAAAAGAElAADIQj0AgKJDeAE=&rs=AOn4CLCLVNMqcLX40_AOmJWixsBn9wWmAQ"}

	req = req_template{"query", "https://www.youtube.com/shorts/r3VI_rAxivM"}
	tests[req] = respond_template{"Najdłuższy SKOK w Minecraft #shorts", 
	"Jam jest Jakub", "https://www.youtube.com/watch?v=r3VI_rAxivM", "https://i.ytimg.com/vi/r3VI_rAxivM/2.jpg"}

	for req, respond := range tests {
		r := httptest.NewRecorder()
		req_json, err := json.Marshal(req)
		if err != nil {
			err_handle(err)
		}
		query_request(r, req_json)
		r_arr := strings.Split(r.Body.String()[:len(r.Body.String())-2], "{")
		respond_json, err := json.Marshal(respond)
		if err != nil {
			err_handle(err)
		}

		if len(r_arr) <= 3 {
			if string(respond_json) != "{" + r_arr[2] {
				t.Errorf("Want: %s Gets: %s", string(respond_json), "{" + r_arr[2])
				t.Fail()
			}
		}else{
			if string(respond_json) + "," != "{" + r_arr[2] {
				t.Errorf("Want: %s Gets: %s", string(respond_json) + ",", r_arr[2])
				t.Fail()
			}
		}
	}
}

func TestVideoInfo(t *testing.T) {
	type req_template struct{
		Request string `json:"request"`
		Link string `json:"link"`
	}
	type respond_template struct {
		Video_quality []string  `json:"video_quality"`
		Audio_quality []float64 `json:"audio_quality"`
	}

	tests := make(map[req_template]respond_template)

	req := req_template{"video-info", "https://www.youtube.com/watch?v=lTAXnaWMR3s"}

	vid_quality := []string{"none","144p","240p","360p","480p","720p","1080p"}
	aud_quality := []float64{127.08,125.162,65.106,49.478,47.596,0}
	tests[req] = respond_template{vid_quality, aud_quality} 

	for req, respond := range tests {
		r := httptest.NewRecorder()
		req_json, err := json.Marshal(req)
		if err != nil {
			err_handle(err)
		}
		videoinfo_request(r, req_json)
		respond_json, err := json.Marshal(respond)
		if err != nil {
			err_handle(err)
		}
		if string(respond_json) != r.Body.String() {
			t.Errorf("Want: %s Gets: %s", string(respond_json), r.Body.String())
			t.Fail()
		}
	}
}

func TestDownload(t *testing.T) {
	type req_template struct{
		Request       string
		Link          string
		Video_quality string  `json:"video-quality"`
		Audio_quality float64 `json:"audio-quality"`
		Format        string
	}

	req := req_template{"download", "https://www.youtube.com/watch?v=U3aWAmSpF6E", "1080p", 360.00, "mp4"}
	Mapa.Map = make(map[string]Download_data)

	r := httptest.NewRecorder()
	req_json, err := json.Marshal(req)
	if err != nil {
		err_handle(err)
	}
	download_request(r, req_json)
	url := strings.Split(r.Body.String(), "\"")

	reader, _ := io.Pipe()
	r = httptest.NewRecorder()

	request, _ := http.NewRequest("GET", url[3], reader)
	id := strings.Split(url[3], "/")
	request = mux.SetURLVars(request, map[string]string{"id": id[2]})

	download_link_generator(r, request)
	
	download := exec.Command("yt-dlp", "-f", "bestvideo[ext=mp4][height<=1080p]+bestaudio[ext=m4a][abr<=360]",
	"-o", "-", "https://www.youtube.com/watch?v=U3aWAmSpF6E")
	file, err := download.StdoutPipe()
	if err != nil {
		err_handle(err)
	}
	download.Start()
	file_byte, err := ioutil.ReadAll(file)
	if err != nil {
		err_handle(err)
	}

	check := bytes.Compare(file_byte, r.Body.Bytes())
	if check != 0{
		t.Errorf("Error: Files are not the same")
		t.Fail()
	}
}