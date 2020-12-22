package wk

import (
	"fmt"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/network"
	"github.com/therecipe/qt/webkit"
)

// GetSnapshot get a snapshot for website
func GetSnapshot(url string, width, height int) {
	imgQuality := 50
	imgFormat := "jpg"
	page := webkit.NewQWebPage(nil)

	networkAccessManager := network.NewQNetworkAccessManager(page)
	networkAccessManager.ConnectSslErrors(func(reply *network.QNetworkReply, errors []*network.QSslError) {
		reply.IgnoreSslErrors()
	})
	page.SetNetworkAccessManager(networkAccessManager)

	page.Settings().SetAttribute(webkit.QWebSettings__WebSecurityEnabled, true)

	page.MainFrame().SetScrollBarPolicy(core.Qt__Horizontal, core.Qt__ScrollBarAlwaysOff)
	page.MainFrame().SetScrollBarPolicy(core.Qt__Vertical, core.Qt__ScrollBarAlwaysOff)

	page.ConnectUserAgentForUrl(func(url *core.QUrl) string {
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A"
	})

	qSize := core.NewQSize2(width, height)
	page.SetViewportSize(qSize)

	qURL := core.NewQUrl3(url, core.QUrl__TolerantMode)
	page.MainFrame().Load(qURL)

	page.ConnectLoadFinished(func(bool) {
		defer page.DeleteLaterDefault()
		defer networkAccessManager.DeleteLaterDefault()
		defer qSize.DestroyQSize()
		defer qURL.DestroyQUrl()
		image := gui.NewQImage3(width, height, gui.QImage__Format_RGB888)
		defer image.DestroyQImageDefault()
		painter := gui.NewQPainter()
		defer painter.DestroyQPainter()
		qPaintDevice := gui.NewQPaintDeviceFromPointer(image.Pointer())
		defer qPaintDevice.DestroyQPaintDeviceDefault()
		painter.Begin(qPaintDevice)

		painter.SetRenderHint(gui.QPainter__Antialiasing, true)
		painter.SetRenderHint(gui.QPainter__TextAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__HighQualityAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__SmoothPixmapTransform, true)

		qRegion := gui.NewQRegion2(0, 0, width, height, gui.QRegion__Rectangle)
		defer qRegion.DestroyQRegion()
		page.MainFrame().Render(painter, qRegion)
		painter.End()

		image.Save("test.jpg", imgFormat, imgQuality)

		buff := core.NewQBuffer(nil)
		defer buff.DeleteLaterDefault()
		buff.Open(core.QIODevice__ReadWrite)
		image.Save2(buff, "jpg", 50)
		data := []byte(buff.Data().ConstData())
		fmt.Println(data)
		// res["data"] = data
	})
}

// ClearCaches clear webkit memory cache
func ClearCaches() {
	webkit.QWebSettings_ClearMemoryCaches()
}
