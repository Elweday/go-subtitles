package renderer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/elweday/go-subtitles/pkg/utils"
)

func FFmpegCombineImagesToVideo(frames [][]byte, inputVideoData []byte, aspectRatio string, frameRate int, offset float64) ([]byte, error) {
	inputFile, err := utils.WriteTemp(inputVideoData)
	if err != nil {
		return nil, err
	}
	defer os.Remove(inputFile.Name())

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "image2pipe",
		"-framerate", fmt.Sprintf("%d", frameRate),
		"-video_size", aspectRatio,
		"-i", "pipe:0",
		"-f", "mp4",
		"-i", inputFile.Name(),
		"-filter_complex", fmt.Sprintf("[1:v][0:v]overlay=0:%f[out]", offset), // Overlay images over background video
		"-map", "[out]",
		"-map", "1:a", // Map the audio stream from the input video
		"-c:a", "copy", // Copy the audio stream without re-encoding
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-f", "matroska",
		"-",
	)

	stdinImages, err := cmd.StdinPipe()
	out := []byte{}
	outBuf := bytes.NewBuffer(out)
	cmd.Stdout = outBuf
	if err != nil {
		return nil, fmt.Errorf("error getting stdin pipe for images: %v", err)
	}

	// Start ffmpeg process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting ffmpeg: %v", err)
	}

	// Write images to stdin
	for _, imgData := range frames {
		_, err := stdinImages.Write(imgData)
		if err != nil {
			return nil, fmt.Errorf("error writing image data to stdin: %v", err)
		}
	}

	// Close stdin for images to signal end of input
	if err := stdinImages.Close(); err != nil {
		return nil, fmt.Errorf("error closing stdin for images: %v", err)
	}

	// Wait for ffmpeg to finish
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error waiting for ffmpeg: %v", err)
	}

	return outBuf.Bytes(), nil
}

func FFmpegExtractAudio(videoBytes []byte) ([]byte, error) {
	// Create pipes for input and output
	reader := bytes.NewReader(videoBytes)
	writer := bytes.NewBuffer(nil)

	// Build ffmpeg command with pipes
	cmd := exec.Command("ffmpeg", "-i", "-", "-vn", "-acodec", "copy", "-")
	cmd.Stdin = reader
	cmd.Stdout = writer    // Pipe output to a buffer
	cmd.Stderr = os.Stderr // Capture ffmpeg errors

	// Run ffmpeg command
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Get extracted audio from the buffer
	audioBytes := writer.Bytes()

	return audioBytes, nil
}

// GetVideoDimensions returns the width and height of a video file as integers
func FFmpegGetVideoDimensions(videoData []byte) (int, int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		"-",
	)

	cmd.Stdin = bytes.NewReader(videoData)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("error running ffprobe: %v", err)
	}

	dimensions := strings.Split(strings.TrimSpace(string(output)), "x")
	if len(dimensions) != 2 {
		return 0, 0, fmt.Errorf("unexpected dimensions format: %s", string(output))
	}

	width, err := strconv.Atoi(dimensions[0])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing width: %v", err)
	}

	height, err := strconv.Atoi(dimensions[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing height: %v", err)
	}

	return width, height, nil
}
