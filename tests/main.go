package main
import (
	"github.com/arrowsio/twitchcap"
	"fmt"
	"os"
	"github.com/arrowsio/twitchcap/m3u"
)

func main() {
	m3u.Import("http://video7.dfw01.hls.ttvnw.net/hls-826098/sodapoppin_18633978064_375348156/chunked/py-index-live.m3u8?token=id=1079323320902777181,bid=18633978064,exp=1451872161,node=video7-1.dfw01.hls.justin.tv,nname=video7.dfw01,fmt=chunked&sig=4665847160e99bed13373dbaae11715aa17d509e")
	//	fmt.Println(err)
	//	fmt.Println(pl)
	c := twitchcap.New()
	if err := c.CaptureStream("sodapoppin"); err != nil {
		panic(err)
	}
	if err := c.FindStream(twitchcap.R_Source); err != nil {
		panic(err)
	}
	fmt.Println("Starting download...")
	file, err := os.OpenFile("/tmp/vod.ts", os.O_APPEND | os.O_RDWR | os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	buf, _ := c.Download(30)
	for b := range buf {
		file.Write(b)
	}
}