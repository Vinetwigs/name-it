package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		OnStartup:        app.startup,
		Title:            "NAME IT",
		Width:            560,
		Height:           720,
		MinWidth:         480,
		MinHeight:        620,
		BackgroundColour: &options.RGBA{R: 12, G: 8, B: 16, A: 255},
		Frameless:        false,
		Assets:           assets,
		Bind:             []any{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
