package Config

type storage struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	Timeout int    `json:"timeout"`
}
