package twitchcap
import (
	"encoding/json"
	"errors"
	"time"
	"fmt"
)

type accessToken struct {
	Token             string
	Sig               string
	Mobile_restricted bool
	Error             string
}

type Capture struct {
	streamer   string
	servers    []downloadServer
	token      *accessToken
	files      chan tsFile
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
	c.servers = make([]downloadServer, 0)
	var err error = nil
	uri := "http://usher.twitch.tv/api/channel/hls/" +
	c.streamer + ".m3u8?allow_spectre=true&token=" +
	c.token.Token + "&player=twitchweb&sig=" + c.token.Sig + "&allow_source=true"
	if body, err := read(uri); err == nil {
		//Get available sources
		matches := anyUrl.FindAllString(string(body), -1)
		for _, uri := range matches {
			c.servers = append(c.servers, resolve(uri))
		}
		if len(c.servers) <= 0 {
			return errors.New("No sutable download stream found.")
		}
	}
	return err
}

func (c *Capture) FindFiles(count, rank int) error {
	var server *downloadServer = nil
	for _, ds := range c.servers {
		if ds.Rank == rank {
			server = ds
		}
	}
	if server == nil {
		return errors.New("Could not find a server to stream from at the 'rank' you wanted. Did you use 1-4?")
	}
	c.files = make(chan tsFile, count)
	go c.seek(server, count)
	return nil
}

func (c *Capture) seek(server *downloadServer, count int) {
	defer close(c.files)
	sent, errs := 0, 0
	for sent < count && errs < 3 {
		if body, err := server.readPlaylist(); err == nil {
			for _, s := range simpleTs.FindAllString(string(body), -1) {
				if _, y := c.downloaded[s]; y || sent >= count {
					continue
				}
				tsf := tsFile{Name:s, Location:server.Base + server.Type + "/" + s}

				c.files <- tsf
				c.downloaded[tsf.Name] = struct{}{}
				sent++
			}
		} else {
			fmt.Println(err)
			errs++
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Capture) Download() chan []byte {
	dc := make(chan []byte)
	go func() {
		buf := make([]byte, 0)
		for tsf := range c.files {
			if bytes, err := tsf.download(); err == nil {
				buf = append(buf, bytes...)
			}
		}
		dc <- buf
		close(dc)
	}()
	return dc
}