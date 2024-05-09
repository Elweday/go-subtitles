package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
