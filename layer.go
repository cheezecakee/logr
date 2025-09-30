package logr

import "strings"

type Layer string

const (
	LayerHTTP Layer = "HTTP"
	LayerDB   Layer = "DB"
	LayerCORE Layer = "CORE"
)

func (l Layer) String() string {
	return string(l)
}

func RegisterLayer(name string) Layer {
	return Layer(strings.ToUpper(name))
}
