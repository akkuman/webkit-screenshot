package wk

import (
	"fmt"
	"os"
	"time"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/network"
	"github.com/therecipe/qt/webkit"
	"github.com/therecipe/qt/widgets"
)

// ScreenshotConfig screenshot config
type ScreenshotConfig struct {
	URL     string
	Width   int
	Height  int
	Quality int
	Format  string
	UA      string
	Timeout time.Duration
}

// FinishCallbackFunc will pass the screenshot data to the callback func when finish screenshot
type FinishCallbackFunc func([]byte)

// ScreenshotObject qt object
type ScreenshotObject struct {
	core.QObject

	_ func(config ScreenshotConfig, finishCallbacks []FinishCallbackFunc) `signal:"StartScreenshot"`
}

// NetworkAccessManager QNetworkAccessManager with timeout
// reference:
//   - phantomjs/src/networkaccessmanager.cpp:createRequest
//   - phantomjs/src/networkaccessmanager.cpp:handleTimeout
//   - https://www.cnblogs.com/apocelipes/p/9361690.html
type NetworkAccessManager struct {
	network.QNetworkAccessManager

	_ func() `constructor:"init"`

	Timeout time.Duration

	errorHandlers []func(code network.QNetworkReply__NetworkError)
}

func (m *NetworkAccessManager) init() {
	m.ConnectCreateRequest(m.createRequest)
	m.ConnectSslErrors(m.sslErrors)
}

// createRequest connect CreateRequest to implement timeout abort feature
func (m *NetworkAccessManager) createRequest(op network.QNetworkAccessManager__Operation, originalReq *network.QNetworkRequest, outgoingData *core.QIODevice) *network.QNetworkReply {
	reply := m.QNetworkAccessManager.CreateRequestDefault(op, originalReq, outgoingData)
	// Use the registered error handler to handle QNetworkReply__NetworkError when they occur
	reply.ConnectError2(func(code network.QNetworkReply__NetworkError) {
		for i := range m.errorHandlers {
			m.errorHandlers[i](code)
		}
	})
	if reply != nil {
		go func() {
			if m.Timeout <= 0 {
				return
			}
			timeout := time.After(m.Timeout)
			select {
			case <-timeout:
				reply.Abort()
			}
		}()
	}
	return reply
}

// sslErrors ignore ssl error
func (m *NetworkAccessManager) sslErrors(reply *network.QNetworkReply, errors []*network.QSslError) {
	reply.IgnoreSslErrors()
}

// registerErrorHandler register handler which hanlde QNetworkReply__NetworkError
func (m *NetworkAccessManager) registerErrorHandler(handler func(code network.QNetworkReply__NetworkError)) {
	m.errorHandlers = append(m.errorHandlers, handler)
}

// NewNetworkAccessManagerWithTimeout create a NetworkAccessManager instance from timeout
func NewNetworkAccessManagerWithTimeout(parent core.QObject_ITF, timeout time.Duration) *NetworkAccessManager {
	networkAccessManager := NewNetworkAccessManager(parent)
	networkAccessManager.Timeout = timeout
	return networkAccessManager
}

// Loader the screenshot loader
type Loader struct {
	*ScreenshotObject

	app *widgets.QApplication
}

// NewLoader create a loader
func NewLoader() *Loader {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")

	app := widgets.NewQApplication(len(os.Args), os.Args)

	l := &Loader{NewScreenshotObject(nil), app}

	l.ConnectStartScreenshot(func(config ScreenshotConfig, finishCallbacks []FinishCallbackFunc) {
		l.GetScreenshot(config, finishCallbacks)
	})

	return l
}

// GetScreenshot get a snapshot for website
func (l *Loader) GetScreenshot(config ScreenshotConfig, finishCallbacks []FinishCallbackFunc) {
	url := config.URL
	width := config.Width
	height := config.Height
	imgQuality := config.Quality
	imgFormat := config.Format
	userAgent := config.UA
	timeout := config.Timeout
	page := webkit.NewQWebPage(nil)
	// indicate whether an error occurred that req is canceled
	isCancelReqError := false

	networkAccessManager := NewNetworkAccessManagerWithTimeout(page, timeout)
	networkAccessManager.registerErrorHandler(func(code network.QNetworkReply__NetworkError) {
		if code == network.QNetworkReply__OperationCanceledError {
			isCancelReqError = true
			if finishCallbacks != nil {
				for i := range finishCallbacks {
					finishCallbacks[i](nil)
				}
			}
		}
	})
	page.SetNetworkAccessManager(networkAccessManager)

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
		// networkAccessManager.DeleteLater() must be executed after page.DeleteLater()
		defer networkAccessManager.DeleteLater()
		defer page.DeleteLater()
		defer qSize.DestroyQSize()
		defer qURL.DestroyQUrl()
		// if req is canceled, return
		if isCancelReqError {
			return
		}
		image := gui.NewQImage3(width, height, gui.QImage__Format_RGB888)
		defer image.DestroyQImageDefault()
		painter := gui.NewQPainter()
		defer painter.DestroyQPainter()
		qPaintDevice := gui.NewQPaintDeviceFromPointer(image.Pointer())
		defer qPaintDevice.DestroyQPaintDeviceDefault()
		painter.Begin(qPaintDevice)

		setPainterRenderHint(painter)

		qRegion := gui.NewQRegion2(0, 0, width, height, gui.QRegion__Rectangle)
		defer qRegion.DestroyQRegion()
		fmt.Println(page.MainFrame().ToHtml())
		page.MainFrame().Render(painter, qRegion)
		painter.End()

		image.Save("test.jpg", imgFormat, imgQuality)

		buff := core.NewQBuffer(nil)
		defer buff.DeleteLater()
		buff.Open(core.QIODevice__ReadWrite)
		image.Save2(buff, imgFormat, imgQuality)
		data := []byte(buff.Data().ConstData())
		// synchronous call the finish callback function
		if finishCallbacks != nil {
			for i := range finishCallbacks {
				finishCallbacks[i](data)
			}
		}
	})
}

// Exec execute qt app main event loop
func (l *Loader) Exec() {
	l.app.Exec()
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

// setPainterRenderHint set RenderHint for painter
func setPainterRenderHint(painter *gui.QPainter) {
	painter.SetRenderHint(gui.QPainter__Antialiasing, true)
	painter.SetRenderHint(gui.QPainter__TextAntialiasing, true)
	painter.SetRenderHint(gui.QPainter__HighQualityAntialiasing, true)
	painter.SetRenderHint(gui.QPainter__SmoothPixmapTransform, true)
}
