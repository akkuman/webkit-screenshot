package main

import (
	"os"

	"github.com/akkuman/webkit-snapshot/wk"
	"github.com/therecipe/qt/widgets"
)

func main() {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")

	app := widgets.NewQApplication(len(os.Args), os.Args)

	wk.GetSnapshot("https://www.baidu.com", 1920, 1080)

	app.Exec()
}
