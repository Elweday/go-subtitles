package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/api/option"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/types"
)

func TranscibeFromStorage(client *speech.Client, gcsURI string) ([]types.Word, error) {
	ctx := context.Background()

	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:              speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz:       16000,
			LanguageCode:          "en-US",
			EnableWordTimeOffsets: true,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	}

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	words := []types.Word{}

	for _, result := range resp.Results {
		// Only first alternative (Top Ranked Result) is considered
		for _, w := range result.Alternatives[0].Words {
			start := float64(w.EndTime.Seconds) + float64(w.EndTime.Nanos)*1e-9
			duration := float64(w.EndTime.Seconds) + float64(w.EndTime.Nanos)*1e-9 - start
			w := types.Word{
				Value:    w.Word,
				Time:     start,
				Duration: duration,
			}
			words = append(words, w)
		}
	}
	return words, nil
}

type GcpIOHandler struct {
	BucketName   string
	InputObject  string
	OutputObject string
	CREDS        []byte
}

func NewGcpHandler() *GcpIOHandler {
	return &GcpIOHandler{
		BucketName:   os.Getenv("SUBTITLES_GCP_BUCKET"),
		InputObject:  os.Getenv("SUBTITLES_GCP_INPUT_OBJECT"),
		OutputObject: os.Getenv("SUBTITLES_GCP_OUTPUT_OBJECT"),
		CREDS:        []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
	}

}

func (handler *GcpIOHandler) SaveVideo(b []byte) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	reader := bytes.NewReader(b)
	// Get Google Cloud Storage bucket
	bucket := client.Bucket(handler.BucketName)

	// Create new object
	obj := bucket.Object(handler.OutputObject)

	// Write content of local file to GCS object
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return fmt.Errorf("failed to write to object: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	log.Printf("File %s uploaded to gs://%s/%s\n", handler.BucketName, handler.BucketName, handler.OutputObject)
	return nil
}

func (handler *GcpIOHandler) Read() (vid *renderer.VidoePayload, err error) {

	return &renderer.VidoePayload{}, nil
}

func Test() {

	b := []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	creds := option.WithCredentialsJSON(b)
	client, _ := speech.NewClient(context.Background(), creds)

	words, _ := TranscibeFromStorage(client, "gs://subtitles-demos/1.mp3")

	fmt.Println(words)
}
