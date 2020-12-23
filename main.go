package main

import (
	"os"

	"github.com/akkuman/webkit-screenshot/wk"
	"github.com/therecipe/qt/widgets"
)

func main() {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")

	app := widgets.NewQApplication(len(os.Args), os.Args)
	screenshotObj := wk.NewScreenshotObject(nil)

	go screenshot(screenshotObj, "https://www.baidu.com")

	app.Exec()
}

func screenshot(obj *wk.ScreenshotObject, url string) {
	config := wk.ScreenshotConfig{
		ID:      "xxxx",
		URL:     url,
		Width:   1920,
		Height:  1080,
		Quality: 50,
		Format:  "jpg",
		UA:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
	}
	obj.StartScreenshot(config)
}
