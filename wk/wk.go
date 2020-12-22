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
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A"
	page := webkit.NewQWebPage(nil)

	networkAccessManager := network.NewQNetworkAccessManager(page)
	networkAccessManager.ConnectSslErrors(func(reply *network.QNetworkReply, errors []*network.QSslError) {
		reply.IgnoreSslErrors()
	})
	page.SetNetworkAccessManager(networkAccessManager)

	page.Settings().SetAttribute(webkit.QWebSettings__WebSecurityEnabled, true)

	setAttributes(page.Settings())

	page.MainFrame().SetScrollBarPolicy(core.Qt__Horizontal, core.Qt__ScrollBarAlwaysOff)
	page.MainFrame().SetScrollBarPolicy(core.Qt__Vertical, core.Qt__ScrollBarAlwaysOff)

	page.ConnectUserAgentForUrl(func(url *core.QUrl) string {
		return userAgent
	})

	qSize := core.NewQSize2(width, height)
	page.SetViewportSize(qSize)

	qURL := core.NewQUrl3(url, core.QUrl__TolerantMode)
	page.MainFrame().Load(qURL)

	page.ConnectLoadFinished(func(bool) {
		defer networkAccessManager.DeleteLater()
		defer page.DeleteLater()
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
		defer buff.DeleteLater()
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

// setAttributes sets web page attributes
func setAttributes(settings *webkit.QWebSettings) {
	// Specifies whether images are automatically loaded in web pages.
	settings.SetAttribute(webkit.QWebSettings__AutoLoadImages, true)
	// Specifies whether QtWebkit will try to pre-fetch DNS entries to speed up browsing.
	settings.SetAttribute(webkit.QWebSettings__DnsPrefetchEnabled, true)
	// Enables or disables the running of JavaScript programs.
	settings.SetAttribute(webkit.QWebSettings__JavascriptEnabled, true)
	// Specifies whether JavaScript programs can open new windows.
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanOpenWindows, false)
	// Specifies whether JavaScript programs can close windows.
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanCloseWindows, false)
	// Specifies whether JavaScript programs can read or write to the clipboard.
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanAccessClipboard, false)
	settings.SetAttribute(webkit.QWebSettings__LocalContentCanAccessFileUrls, true)
	settings.SetAttribute(webkit.QWebSettings__LocalContentCanAccessRemoteUrls, true)
	settings.SetAttribute(webkit.QWebSettings__SiteSpecificQuirksEnabled, true)
	settings.SetAttribute(webkit.QWebSettings__PrivateBrowsingEnabled, true)

	settings.SetAttribute(webkit.QWebSettings__PluginsEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__JavaEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__WebGLEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__WebAudioEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__NotificationsEnabled, false)

	settings.SetAttribute(webkit.QWebSettings__Accelerated2dCanvasEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__AcceleratedCompositingEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__TiledBackingStoreEnabled, false)

	settings.SetAttribute(webkit.QWebSettings__LocalStorageEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__OfflineStorageDatabaseEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__OfflineWebApplicationCacheEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__WebSecurityEnabled, true)
}
