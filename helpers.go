package main

import (
	"golang.org/x/image/font"
    "golang.org/x/image/math/fixed"
	types "github.com/elweday/go-subtitles/utils"
)

func WrapLines(input []types.Word, face font.Face, delimiter string, maxLines int, totalWidth fixed.Int26_6) [][]types.Word {
    var chunks [][]types.Word
    var currentChunk []types.Word
    currentWidth := fixed.I(0)
    lines := 0

    for _, word := range input {
        // Calculate the width of the word.
        wordWidth := font.MeasureString(face, word.Value)

        // Calculate the width of the delimiter.
        delimiterWidth := font.MeasureString(face, delimiter)

        // If adding this word (plus delimiter if any) exceeds the total width, start a new line.
        if currentWidth+wordWidth+delimiterWidth > totalWidth {
            chunks = append(chunks, currentChunk)
            currentChunk = nil
            currentWidth = fixed.I(0)
            lines++
        }

        // If the number of lines exceeds the maxLines, start a new chunk.
        if lines >= maxLines {
            chunks = append(chunks, currentChunk)
            currentChunk = nil
            currentWidth = fixed.I(0)
            lines = 0
        }

        currentChunk = append(currentChunk, word)
        currentWidth += wordWidth

        // Add the width of the delimiter if any.
        currentWidth += delimiterWidth
    }

    if len(currentChunk) > 0 {
        chunks = append(chunks, currentChunk)
    }

    return chunks
}