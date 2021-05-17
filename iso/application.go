package iso

type Application struct {
	Name       string `json:"name"`
	Executable string `json:"executable"`
	Port       string `json:"port"`
	Route      string `json:"route"`
	Scheme	   string `json:"scheme"`
}

type Applications struct {
	Applications []Application `json:"applications"`
}