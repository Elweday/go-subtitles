package gcp

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
	"github.com/elweday/go-subtitles/src/types"
	"github.com/joho/godotenv"
)

func Upload(bucketName string, objectName string, b []byte) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	reader := bytes.NewReader(b)
	// Get Google Cloud Storage bucket
	bucket := client.Bucket(bucketName)

	// Create new object
	obj := bucket.Object(objectName)

	// Write content of local file to GCS object
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		log.Fatalf("Failed to write to object: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("Failed to close writer: %v", err)
	}

	log.Printf("File %s uploaded to gs://%s/%s\n", bucketName, objectName, objectName)
}

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

func TestEndPoint() {

	godotenv.Load(".env")
	b := []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	creds := option.WithCredentialsJSON(b)

	client, _ := speech.NewClient(context.Background(), creds)

	words, _ := TranscibeFromStorage(client, "gs://subtitles-demos/1.mp3")

	fmt.Println(words)
}
