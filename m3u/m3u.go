package m3u
import (
	"net/http"
	"bufio"
	"errors"
	"strings"
	"strconv"
	"regexp"
)

var (
	findBase = regexp.MustCompile(`^(http:\/\/.+\/)(.*?)(\?|$)`)
)

type M3U struct {
	URL       string
	base      string
	Length    float32
	Parts     []M3UPart
	PlayLists []M3UStream
}

type M3UPart struct {
	Name   string
	File   string
	Path   string
	Length float32
}

type M3UStream struct {
	Video      string
	Name       string
	Resolution string
	Location   string
	Rank       int
}

func Import(url string) (*M3U) {
	return &M3U{URL: url, base: findBase.FindStringSubmatch(url)[1]}
}

func (this *M3U) Read() error {
	this.Parts = this.Parts[:0]
	this.PlayLists = this.PlayLists[:0]
	resp, err := http.Get(this.URL)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(resp.Body)
	if !(scanner.Scan() && scanner.Text() == "#EXTM3U") {
		return errors.New("M3U not valid.")
	}
	// Interpret
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			if len(line) < 8 {
				continue
			}
			if strings.HasPrefix(line, "#EXTINF:") {
				this.extINF(scanner)
			} else if strings.HasPrefix(line, "#EXT-X-MEDIA:") {
				this.media(scanner)
			}
		}
	}
	return nil
}

func (this *M3U) extINF(scanner *bufio.Scanner) {
	line := scanner.Text()
	part := M3UPart{}
	i := strings.Index(line, ",")
	if f, err := strconv.ParseFloat(line[8:i], 32); err == nil {
		part.Length = float32(f)
		part.Name = line[i:]
		scanner.Scan()
		part.File = scanner.Text()
		part.Path = this.base + part.File
		this.Parts = append(this.Parts, part)
		this.Length += part.Length
	}
}

func (this *M3U) media(scanner *bufio.Scanner) {
	line := scanner.Text()
	buf := make(map[string]string)
	scanner.Scan()
	l2 := scanner.Text()
	scanner.Scan()
	toMap(line[strings.Index(line, ":") + 1:] + "," + l2[strings.Index(l2, ":") + 1:], buf)
	rank := 0
	switch buf["VIDEO"] {
	case "chunked":
		rank = 1
	case "high":
		rank = 2
	case "medium":
		rank = 3
	case "low":
		rank = 4
	case "mobile":
		rank = 5
	}
	this.PlayLists = append(this.PlayLists, M3UStream{
		Video:buf["VIDEO"],
		Name:buf["NAME"],
		Resolution:buf["RESOLUTION"],
		Location: scanner.Text(),
		Rank:rank,
	})
}

func toMap(data string, buf map[string]string) map[string]string {
	data += ","
	name := ""
	skip, quotes := false, false
	from := 0
	for i, c := range data {
		if c == '"' {
			quotes = true
			skip = !skip
		}
		if skip {
			continue
		}
		if c == '=' {
			name = data[from:i]
			from = i + 1
		}
		if c == ',' {
			if quotes {
				buf[name] = data[from+1:i-1]
				quotes = false
			} else {
				buf[name] = data[from:i]
			}
			from = i + 1
		}
	}
	return buf
}