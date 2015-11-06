package main
import (
	"net/http"
	"time"
	"arrows.io/cap/twitch"
	"io/ioutil"
	"encoding/json"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", download)
	http.ListenAndServe(":8080", mux)
}

type downloaded struct {
	Size float32
	Time time.Duration
}

func download(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if len(user) == 0 {
		http.Error(w, "No user provided.", 500)
		return
	}
	start := time.Now()
	cap, err := twitch.NewCapture(user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	cap.FindFiles(7)
	buf := <-cap.Download()
	dur := time.Now().Sub(start)
	ioutil.WriteFile("/tmp/testing.ts", buf, 0644)
	if data, err := json.Marshal(&downloaded{float32(len(buf)) / 2048, dur}); err == nil {
		w.Write(data)
		return
	}
	http.Error(w, "Could not marshal struct :(", 500)
}