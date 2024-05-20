package main

import (
	"fmt"
	"log"
	"os"

	"github.com/elweday/go-subtitles/pkg/handlers"

	"encoding/json"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func renderSubtitles(handler handlers.IOHandler) error {
	vid, err := handler.Read()
	fmt.Println(err)
	err = vid.RenderWithSubtitles()
	if err != nil {
		return err
	}
	err = handler.SaveVideo(vid.OutputVideo)
	if err != nil {
		return err
	}

	return nil

}

func main() {
	godotenv.Load()

	runEnv := os.Getenv("SUBTITLES_RUN_ENVIRONMENT")
	if runEnv == "LOCAL" {
		handler := &handlers.LocalIOHandler{
			TranscriptPath: os.Getenv("SUBTITLES_TRANSCRIPT_PATH"),
			ConfigPath:     os.Getenv("SUBTITLES_CONFIG_PATH"),
			InputVideoPath: os.Getenv("SUBTITLES_INPUT_VIDEO_PATH"),
			OutputPath:     os.Getenv("SUBTITLES_OUTPUT_VIDEO_PATH"),
		}
		err := renderSubtitles(handler)
		if err != nil {
			log.Fatal(err)
		}

	} else if runEnv == "GCP" {
		functions.HTTP("render", cloudFunctionHandler)
		port := "8080"
		if envPort := os.Getenv("SUBTITLES_FUNCION_PORT"); envPort != "" {
			port = envPort
		}
		if err := funcframework.Start(port); err != nil {
			log.Fatalf("funcframework.Start: %v\n", err)
		}

	}
}

/*
* Authorization: Bearer <token>
* Body: { "bucketName": string "doc": string "projectID": string }
 */
func cloudFunctionHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	handler := &handlers.GcpIOHandler{}
	if err := json.NewDecoder(r.Body).Decode(handler); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	if err := renderSubtitles(handler); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "success")

}
