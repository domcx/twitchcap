package twitchcap
import (
	"errors"
	"fmt"
	"github.com/arrowsio/twitchcap/m3u"
)

type accessToken struct {
	Token             string
	Sig               string
	Mobile_restricted bool
	Error             string
}

type Capture struct {
	source          string
	token           *accessToken
	downloaded      map[string]struct{}
	playlist        *m3u.M3U
	isVod           bool
	stopDownloading bool
}

func New() *Capture {
	return &Capture{downloaded: make(map[string]struct{})}
}

func (c *Capture) CaptureStream(streamer string) error {
	c.source = streamer
	return c.init("https://api.twitch.tv/api/channels/" + streamer + "/access_token?as3=t")
}

func (c *Capture) CaptureVod(vod string) error {
	c.source = vod
	c.isVod = true
	return c.init("https://api.twitch.tv/api/vods/" + vod + "/access_token?as3=t")
}

func (c *Capture) init(url string) error {
	c.token = &accessToken{}
	if err := readJson(url, &c.token); err != nil {
		return err
	}
	if c.token.Token == "" {
		return errors.New("Unable to obtain streaming token.")
	}
	return nil
}

func (c *Capture) FindStream(rank int) (error) {
	var uri string
	if c.isVod {
		uri = fmt.Sprintf("http://usher.twitch.tv/vod/%v?allow_spectre=true&nauth=%v&player=twitchweb&nauthsig=%v&allow_source=true", c.source, c.token.Token, c.token.Sig)
	} else {
		uri = fmt.Sprintf("http://usher.twitch.tv/api/channel/hls/%v.m3u8?token=%v&allow_source=true&player=twitchweb&sig=%v&allow_spectre=true", c.source, c.token.Token, c.token.Sig)
	}
	m := m3u.Import(uri)
	if err := m.Read(); err != nil {
		return err
	}
	for _, stream := range m.PlayLists {
		if stream.Rank == rank {
			c.playlist = m3u.Import(stream.Location)
			return nil
		}
	}
	return errors.New("Stream with that ranking not found.")
}

func (c *Capture) Stop() {
	c.stopDownloading = true
}

func (c *Capture) Download(timeout float32) (chan []byte, chan error) {
	buf := make(chan []byte, 15)
	errBuf := make(chan error, 15)
	c.stopDownloading = false
	go func() {
		defer close(buf)
		var elapsed float32 = 0.0
		for elapsed < timeout && !c.stopDownloading {
			if err := c.playlist.Read(); err != nil {
				errBuf <- err
			}
			for _, part := range c.playlist.Parts {
				if elapsed >= timeout || c.stopDownloading {
					break
				}
				if err := c.playlist.Read(); err != nil {
					errBuf <- err
					break
				}
				if data, err := readRaw(part.Path); err == nil {
					buf <- data
				} else {
					errBuf <- err
				}
				elapsed += part.Length
				fmt.Println(elapsed)
			}
		}
	}()
	return buf, errBuf
}
