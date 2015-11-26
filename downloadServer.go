package twitchcap

type downloadServer struct {
	Name, Type, Playlist, Base string
	Rank                       int
}

func resolve(url string) (downloadServer) {
	st := urlInfo.FindStringSubmatch(url)
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

func (ds *downloadServer) readPlaylist() ([]byte, error) {
	return read(ds.Playlist)
}
