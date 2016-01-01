# twitchcap
Capture twitch moments

```go

  user := "STREAMER"
  cap, err := twitchcap.NewCapture(user)
  if err != nil {
    panic(err)
  }
  cap.FindFiles(7) // Number of .ts files you want it to record in succession.
  buf := <-cap.Download() // Wait for the download channel to close, giving us the bytes of the raw video.
  ioutil.WriteFile("myvideo.ts", buf, 0644) //Now do whatever you want with the bytes?
```
