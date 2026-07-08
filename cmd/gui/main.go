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
		OnStartup:  app.Startup,
		OnShutdown: app.shutdown,
		// GPU acceleration is intentionally left ENABLED. It was previously disabled
		// (WebviewGpuIsDisabled: true) to avoid a brief DWM hardware-cursor stall
		// during GPU compositor init on startup, but software rendering does not
		// re-rasterize when the window moves between monitors with different DPI
		// scaling -> blurry text on the second monitor. Crisp cross-DPI rendering is
		// the more important trade-off.
		Windows: &windows.Options{
			WebviewGpuIsDisabled: false,
		},
		Bind: []any{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}
