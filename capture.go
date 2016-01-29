package twitchcap
import (
	"errors"
	"fmt"
	"github.com/pranked/twitchcap/m3u"
	"strconv"
	"strings"
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

// Find the playlist with the stream with the required rank
// PLEASE NOTE THAT IF YOUR NETWORK CANNOT HANDLE THE DOWNLOAD, SOME CHUNKS OF VIDEO WILL BE LOST.
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

// Stop downloading
func (c *Capture) Stop() {
	c.stopDownloading = true
}

// Download from the stream until video length EQUALS OR EXCEEDS the timeout requested.
func (c *Capture) Download(timeout float32) (chan []byte, chan error) {
	buf := make(chan []byte, 5)
	parts := make(chan m3u.M3UPart, 25)
	errBuf := make(chan error, 5)
	c.stopDownloading = false
	go func() {
		defer close(buf)
		defer close(errBuf)
		// Find the ts files
		go func() {
			defer close(parts)
			var timeDownloaded float32 = 0
			var last uint32 = 0
			for timeDownloaded < timeout && !c.stopDownloading {
				if err := c.playlist.Read(); err != nil {
					// In this instance it is most likely that the streamer is no longer streaming, or a VOD no longer exists
					errBuf <- err
					return
				}
				for _, part := range c.playlist.Parts {
					if timeDownloaded >= timeout || c.stopDownloading {
						return
					}
					// Make sure we don't already have this file (uses the index-0000000000-, as uint32 in case of overflow)
					s := part.File[6:16]
					if c.isVod {
						s = strings.Split(part.File, "end_offset=")[1]
					}
					if u, err := strconv.ParseUint(s, 10, 32); err != nil {
						errBuf <- err
					} else {
						//If we already have this file, skip it
						if last != 0 && uint32(u) <= last {
							continue
						}
						last = uint32(u)
					}
					parts <- part
					timeDownloaded += part.Length
				}
			}
		}()
		for part := range parts {
			// Read the input file.
			if data, err := readRaw(part.Path); err == nil {
				buf <- data
			} else {
				errBuf <- err
			}
		}
	}()
	return buf, errBuf
}
