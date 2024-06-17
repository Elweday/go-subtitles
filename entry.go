package renderSubtitles

import (
	"fmt"
	"io"

	"github.com/elweday/go-subtitles/pkg/handlers"

	"encoding/json"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func renderSubtitles(handler handlers.IOHandler) error {
	vid, err := handler.Read()
	if err != nil {
		return err
	}
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

func init() {
	functions.HTTP("RenderSubtitles", RenderSubtitles)
}

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func RenderSubtitles(w http.ResponseWriter, r *http.Request) {
	handler := &handlers.EndPointHandler{}

	if err := json.NewDecoder(r.Body).Decode(handler); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body, _ := io.ReadAll(r.Body)
		fmt.Fprint(w, "error parsing request body: "+err.Error(), "\n", string(body))
		return
	}

	if handler.InputVideo == nil || handler.Transcript == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid request body, missing inputVideo or transcript")
		return
	}

	// render subtitle start
	vid, err := handler.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "couldn't read body: "+err.Error())
		return

	}
	err = vid.RenderWithSubtitles()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "couldn't render subtitles for video: "+err.Error())
		return
	}

	w.Header().Add("Content-Type", "video/mp4")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(vid.OutputVideo))
	// fmt.Fprint(w, handler.Out)
}
