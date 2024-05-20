package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/api/option"

	"cloud.google.com/go/firestore"
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
	Doc          string
	ProjectID    string
	CREDS        []byte
}

func NewGcpHandler() *GcpIOHandler {
	return &GcpIOHandler{
		BucketName: os.Getenv("SUBTITLES_GCP_BUCKET"),
		Doc:        os.Getenv("SUBTITLES_GCP_DOC"),
		ProjectID:  os.Getenv("SUBTITLES_GCP_PROJECT_ID"),
		CREDS:      []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
	}
}

func (handler *GcpIOHandler) Auth() option.ClientOption {
	return option.WithCredentialsJSON(handler.CREDS)
}

func (handler *GcpIOHandler) SaveVideo(b []byte) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, handler.Auth())
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

	log.Printf("File uploaded to gs://%s/%s\n", handler.BucketName, handler.OutputObject)
	return nil
}

func (handler *GcpIOHandler) ReadInput() ([]byte, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, handler.Auth())
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	objBytes := []byte{}
	inReader := bytes.NewBuffer(objBytes)
	// Get Google Cloud Storage bucket
	bucket := client.Bucket(handler.BucketName)

	obj := bucket.Object(handler.InputObject)
	storageReader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create newReader: %v", err)
	}

	if _, err := io.Copy(inReader, storageReader); err != nil {
		return nil, fmt.Errorf("failed to write to object: %v", err)
	}
	if err := storageReader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	log.Printf("File read from gs://%s/%s\n", handler.BucketName, handler.InputObject)
	return inReader.Bytes(), nil
}

func (handler *GcpIOHandler) Read() (vid *renderer.VidoePayload, err error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, handler.ProjectID, handler.Auth())
	if err != nil {
		return vid, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	doc := client.Doc(handler.Doc)

	docsnap, err := doc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %v", err)
	}

	if err := docsnap.DataTo(vid); err != nil {
		return nil, fmt.Errorf("failed to decode document: %v", err)
	}
	log.Printf("Document read from firestore: %s\n", handler.Doc)

	handler.InputObject = vid.InputVideoObj
	handler.OutputObject = vid.OutputVideoObj
	videoBytes, err := handler.ReadInput()
	vid.InputVideo = videoBytes
	if err != nil {
		return nil, fmt.Errorf("failed to read video: %v", err)
	}
	w, h, err := renderer.FFmpegGetVideoDimensions(videoBytes)
	vid.Opts.Width = w
	vid.Opts.Height = h

	return vid, nil
}

func Test() {

	b := []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	// TODO: Change authenrication method
	creds := option.WithCredentialsJSON(b)
	client, _ := speech.NewClient(context.Background(), creds)

	words, _ := TranscibeFromStorage(client, "gs://subtitles-demos/1.mp3")

	fmt.Println(words)
}
