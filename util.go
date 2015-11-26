package twitchcap
import (
"io/ioutil"
"net/http"
	"regexp"
)

var (
	anyUrl = regexp.MustCompile("http.?:\\/\\/.+")
	simpleTs = regexp.MustCompile(".+\\.ts")
	urlInfo = regexp.MustCompile(`^(.+)/(.+)/py-index-live.m3u8?.+nname=(.+)[,|$].+`)
)

func read(url string) ([]byte, error) {
	var err error = nil
	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return nil, err
}
