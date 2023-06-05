package serializers

import "encoding/json"

type JsonSerializer[T any] struct {
}

func (s *JsonSerializer[T]) Serialize(i T) ([]byte, error) {
	return json.Marshal(i)
}

func (s *JsonSerializer[T]) Deserialize(b []byte, i *T) error {
	return json.Unmarshal(b, i)
}
