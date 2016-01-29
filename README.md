# twitchcap
Capture twitch moments

```go
	c := twitchcap.New()
	//Select source
	if err := c.CaptureVod("34138940"); err != nil {
		panic(err)
	}
	if err := c.FindStream(twitchcap.R_Source); err != nil {
		panic(err)
	}
	fmt.Println("Starting download...")
	//Maybe we save it?
	file, err := os.OpenFile("/tmp/vod.ts", os.O_APPEND | os.O_RDWR | os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	//Download time in seconds, approximate
	buf, _ := c.Download(60)
	//Each part of the stream gets sent here
	for b := range buf {
		file.Write(b)
	}
```
