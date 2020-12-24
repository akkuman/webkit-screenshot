package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/therecipe/qt/webkit"

	"github.com/akkuman/webkit-screenshot/wk"
)

func main() {
	loader := wk.NewLoader()

	url := os.Args[1]

	config := wk.NewScreenshotConfig(url).WithTimeout(30 * time.Second)

	config.RegisterWebFrameHandler(wk.WebFrameHandler(func(frame *webkit.QWebFrame) {
		htmlContent := frame.ToHtml()
		if strings.Contains(htmlContent, "doesn't work properly without JavaScript enabled. Please enable it to continue.") {
			time.Sleep(2 * time.Second)
		}
	}))

	config.RegisterResultHandler(wk.ResultHandler(func(data []byte) {
		fmt.Println(data)
	}))

	go func() {
		data := loader.Screenshot(config)
		ioutil.WriteFile("test.jpg", data, 0644)
	}()

	loader.Exec()
}
