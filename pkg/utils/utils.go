package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"

	"github.com/abdullahdiaa/garabic"
	"github.com/elweday/go-subtitles/pkg/types"

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
	f, err := os.CreateTemp("", uuid.New().String())
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

func ReadAndConvertToFrames(jsonString []byte, frameRate int) ([]types.Word, error) {
	var items = []types.Word{}
	reader := bytes.NewReader(jsonString)

	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&items); err != nil {
		return nil, errors.New("encoding to []Word failed: make sure the file has the appropriate format")
	}

	for i := range items {
		items[i].Frames = int64(math.Round(items[i].Time * float64(frameRate)))
		items[i].Value = garabic.Shape(items[i].Value)
	}

	items = append([]types.Word{}, items...)

	return items, nil
}
