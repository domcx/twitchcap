package twitch
import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"regexp"
	"errors"
	"time"
)

type accessToken struct {
	Token             string
	Sig               string
	Mobile_restricted bool
	Error             string
}

type downloadServer struct {
	Name, Type, Playlist, Base string
	Rank                       int
}

func resolve(url string) (downloadServer) {
	reg := regexp.MustCompile(`^(.+)/(.+)/py-index-live.m3u8?.+nname=(.+)[,|$].+`)
	st := reg.FindStringSubmatch(url)
	ds := downloadServer{Playlist: st[0], Base:st[1] + "/", Type:st[2], Name:st[3]}
	ds.Rank = 0
	switch ds.Type {
	case "chunked":
		ds.Rank = 1
	case "high":
		ds.Rank = 2
	case "medium":
		ds.Rank = 3
	case "low":
		ds.Rank = 4
	}
	return ds
}

func (ds *downloadServer) readPlaylist(client *http.Client) (string, error) {
	body, err := read(ds.Playlist)
	return string(body), err
}

type tsFile struct {
	Name, Location string
}

func (tsf *tsFile) download(client *http.Client) ([]byte, error) {
	return read(tsf.Location)
}

type Capture struct {
	streamer   string
	servers    []downloadServer
	token      *accessToken
	files      chan tsFile
	client     *http.Client
	downloaded map[string]struct{}
}

func NewCapture(streamer string) (*Capture, error) {
	var err error = nil
	if body, err := read("https://api.twitch.tv/api/channels/" + streamer + "/access_token?as3=t"); err == nil {
		tk := accessToken{}
		if err = json.Unmarshal(body, &tk); err == nil {
			if tk.Token == "" {
				return nil, errors.New("User is not streaming.")
			}
			cap := Capture{streamer: streamer, token:&tk}
			cap.downloaded = make(map[string]struct{})
			if err = cap.findServers(); err == nil {
				return &cap, nil
			}
		}
	}
	return nil, err
}

func (c *Capture) findServers() (error) {
	var err error = nil
	uri := "http://usher.twitch.tv/api/channel/hls/" +
	c.streamer + ".m3u8?allow_spectre=true&token=" +
	c.token.Token + "&player=twitchweb&sig=" + c.token.Sig + "&allow_source=true"
	c.servers = make([]downloadServer, 1)
	if body, err := read(uri); err == nil {
		//Get available sources
		reg := regexp.MustCompile("http.?:\\/\\/.+")
		matches := reg.FindAllString(string(body), -1)
		for _, uri := range matches {
			c.servers = append(c.servers, resolve(uri))
		}
		if len(c.servers) <= 0 {
			return errors.New("No sutable download stream found.")
		}
	}
	return err
}

func (c *Capture) FindFiles(count int) {
	var server *downloadServer = nil
	for _, ds := range c.servers {
		if server == nil || ds.Rank < server.Rank {
			server = &ds;
			break
		}
	}
	c.files = make(chan tsFile, count)
	go c.seek(server, count)
}

func (c *Capture) seek(server *downloadServer, count int) {
	defer close(c.files)
	sent, errs := 0, 0
	regTs := regexp.MustCompile(".+\\.ts")
	for sent < count {
		if body, err := server.readPlaylist(c.client); err == nil {
			for _, s := range regTs.FindAllString(string(body), -1) {
				if _, y := c.downloaded[s]; y || sent >= count {
					continue
				}
				tsf := tsFile{Name:s, Location:server.Base + server.Type + "/" + s}
				c.files <- tsf
				c.downloaded[tsf.Name] = struct{}{}
				sent++
			}
		} else {
			errs++
			if errs >= 3 {
				break
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Capture) Download() chan []byte {
	dc := make(chan []byte)
	go func() {
		buf := make([]byte, 0)
		for tsf := range c.files {
			if bytes, err := tsf.download(c.client); err == nil {
				buf = append(buf, bytes...)
			}
		}
		dc <- buf
		close(dc)
	}()
	return dc
}

func read(url string) ([]byte, error) {
	var err error = nil
	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return nil, err
}