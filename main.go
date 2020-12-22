package main

/*
#include <qplugin.h>
Q_IMPORT_PLUGIN(QOffscreenIntegrationPlugin)
Q_IMPORT_PLUGIN(QJpegPlugin)
Q_IMPORT_PLUGIN(QGifPlugin)
*/

import (
	"os"

	"C"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/network"
	"github.com/therecipe/qt/webkit"
	"github.com/therecipe/qt/widgets"
)

func main() {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")

	app := widgets.NewQApplication(len(os.Args), os.Args)

	getSnapshot("https://www.baidu.com", 1920, 1080)

	app.Exec()
}

func getSnapshot(url string, width, height int) {
	page := webkit.NewQWebPage(nil)

	networkAccessManager := network.NewQNetworkAccessManager(page)
	networkAccessManager.ConnectSslErrors(func(reply *network.QNetworkReply, errors []*network.QSslError) {
		reply.IgnoreSslErrors()
	})
	page.SetNetworkAccessManager(networkAccessManager)

	page.Settings().SetAttribute(webkit.QWebSettings__WebSecurityEnabled, true)

	page.MainFrame().SetScrollBarPolicy(core.Qt__Horizontal, core.Qt__ScrollBarAlwaysOff)
	page.MainFrame().SetScrollBarPolicy(core.Qt__Vertical, core.Qt__ScrollBarAlwaysOff)

	page.SetViewportSize(core.NewQSize2(width, height))

	qURL := core.NewQUrl3(url, core.QUrl__TolerantMode)

	page.MainFrame().Load(qURL)

	page.ConnectLoadFinished(func(bool) {
		image := gui.NewQImage3(width, height, gui.QImage__Format_RGB888)
		painter := gui.NewQPainter()
		painter.Begin(gui.NewQPaintDeviceFromPointer(image.Pointer()))

		painter.SetRenderHint(gui.QPainter__Antialiasing, true)
		painter.SetRenderHint(gui.QPainter__TextAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__HighQualityAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__SmoothPixmapTransform, true)

		qRegion := gui.NewQRegion2(0, 0, width, height, gui.QRegion__Rectangle)
		page.MainFrame().Render(painter, qRegion)
		painter.End()

		buff := core.NewQBuffer(nil)
		buff.Open(core.QIODevice__ReadWrite)
		image.Save("test.jpg", "jpg", 50)
	})
}
