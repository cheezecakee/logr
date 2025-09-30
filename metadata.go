package logr

type Metadata struct {
	data map[string]any
}

func NewMetadata() *Metadata {
	return &Metadata{
		data: make(map[string]any),
	}
}

func (m *Metadata) Add(key string, value any) {
	m.data[key] = value
}

func (m *Metadata) Get(key string) (any, bool) {
	value, ok := m.data[key]
	if ok {
		return value, true
	}
	return nil, false
}
