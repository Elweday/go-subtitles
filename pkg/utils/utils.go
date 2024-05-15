package utils

import (
	"os"

	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

func Iff[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

type Number interface {
	constraints.Integer | constraints.Float
}

func WriteTemp(data []byte) (*os.File, error) {
	f, err := os.CreateTemp("/tmp", uuid.New().String())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return nil, err
	}
	return f, nil
}
