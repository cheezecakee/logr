package logr

type Metadata struct {
	Data map[string]any `json:"data"`
}

func NewMetadata() *Metadata {
	return &Metadata{
		Data: make(map[string]any),
	}
}

func (m *Metadata) Add(key string, value any) {
	m.Data[key] = value
}

func (m *Metadata) Get(key string) (any, bool) {
	value, ok := m.Data[key]
	if ok {
		return value, true
	}
	return nil, false
}
