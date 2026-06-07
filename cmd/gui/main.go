package main

import (
	"embed"

	_ "git.um-react.app/um/cli/algo/kgm"
	_ "git.um-react.app/um/cli/algo/kwm"
	_ "git.um-react.app/um/cli/algo/ncm"
	_ "git.um-react.app/um/cli/algo/tm"
	_ "git.um-react.app/um/cli/algo/xiami"
	_ "git.um-react.app/um/cli/algo/ximalaya"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Unlock Music",
		Width:  960,
		Height: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		OnStartup: app.Startup,
		// Disable WebView2 GPU acceleration to avoid a brief DWM hardware-cursor
		// stall during GPU compositor init on startup. The UI is simple, so CPU
		// rendering is fine.
		Windows: &windows.Options{
			WebviewGpuIsDisabled: true,
		},
		Bind: []any{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}
