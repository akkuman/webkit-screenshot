package main

import (
	"fmt"

	"github.com/akkuman/webkit-screenshot/wk"
)

func main() {
	loader := wk.NewLoader()

	go screenshot(loader, "https://www.baidu.com")

	loader.Exec()
}

func screenshot(loader *wk.Loader, url string) {
	config := wk.ScreenshotConfig{
		ID:      "xxxx",
		URL:     url,
		Width:   1920,
		Height:  1080,
		Quality: 50,
		Format:  "jpg",
		UA:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
	}
	dataChan := make(chan []byte)
	var finishCallbacks []wk.FinishCallbackFunc
	finishCallback := wk.FinishCallbackFunc(func(data []byte) {
		dataChan <- data
		close(dataChan)
	})
	finishCallbacks = append(finishCallbacks, finishCallback)
	loader.StartScreenshot(config, finishCallbacks)
	screenshotBytes := <-dataChan
	fmt.Println(screenshotBytes)
}
