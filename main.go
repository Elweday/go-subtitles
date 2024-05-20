package main

import (
	"fmt"
	"time"

	"github.com/elweday/go-subtitles/pkg/handlers"

	"github.com/joho/godotenv"
)

func main() {
	start := time.Now()
	defer since(start)
	godotenv.Load()

	handler := handlers.GetIOHnadler()
	vid, err := handler.Read()
	fmt.Println(err)
	err = vid.RenderWithSubtitles()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = handler.SaveVideo(vid.OutputVideo)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func since(start time.Time) {
	fmt.Printf("\nElapsed time: %s\n", time.Since(start))
}
