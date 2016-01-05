package twitchcap
import (
	"io/ioutil"
	"net/http"
	"encoding/json"
)

const (
	R_Source = 1
	R_High = 2
	R_Medium = 3
	R_Low = 4
	R_Mobile = 5
)

func readRaw(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func readJson(url string, v interface{}) (error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(v)
}