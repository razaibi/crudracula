package encoders

import (
	"github.com/bytedance/sonic"
)

type SonicEncoder struct{}

func Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal is a function type that implements utils.JSONUnmarshal
func Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}
