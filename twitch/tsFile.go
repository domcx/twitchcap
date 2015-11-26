package twitch

type tsFile struct {
	Name, Location string
}

func (tsf *tsFile) download() ([]byte, error) {
	return read(tsf.Location)
}