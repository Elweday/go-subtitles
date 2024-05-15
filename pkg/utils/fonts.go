package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font"
)

func GetFontWeightMapFromGoogle(fontName string, subsets string) (map[string][]byte, error) {
	URL := fmt.Sprintf("https://gwfh.mranftl.com/api/fonts/%s?download=zip&formats=ttf&subsets=%s", fontName, subsets)
	res, err := http.Get(URL)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download font: %s", res.Status)
	}

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	// Read the zip file contents from the buffer
	fontZip := buf.Bytes()

	// Create a new reader for the zip file contents
	zipReader, err := zip.NewReader(bytes.NewReader(fontZip), int64(len(fontZip)))
	if err != nil {
		return nil, err
	}

	// Initialize a map to store font weights
	fontWeightMap := make(map[string][]byte)

	re := regexp.MustCompile(`^.*-(\d+|regular)\.ttf$`)

	// Iterate through each file in the zip
	for _, file := range zipReader.File {
		// Open the file
		fileReader, err := file.Open()

		if err != nil {
			return nil, err
		}

		defer fileReader.Close()

		if re.MatchString(file.Name) {
			fontWeight := re.FindStringSubmatch(file.Name)[1]
			fontWeightBytes, err := io.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}

			fontWeightMap[fontWeight] = fontWeightBytes
		}
	}

	return fontWeightMap, nil
}

func ReadFont(fontBytes []byte, size float64) font.Face {
	f, _ := freetype.ParseFont(fontBytes)

	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
	})
	return face
}

type FontPool struct {
	sync.Pool
	fontSize float64
}

func (p *FontPool) Get(key string) *font.Face {

	fonts := p.Pool.Get().(map[string]font.Face)
	defer p.Pool.Put(fonts)

	font, ok := fonts[key]
	if ok {
		return &font
	}

	return nil
}

func NewPool(m map[string][]byte, fontSize float64) *FontPool {
	return &FontPool{
		fontSize: fontSize,
		Pool: sync.Pool{
			New: func() any {
				items := map[string]font.Face{}
				for k, v := range m {
					face := ReadFont(v, fontSize)
					items[k] = face
				}
				return items
			},
		},
	}
}
