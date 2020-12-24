package wk

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/network"
	"github.com/therecipe/qt/webkit"
	"github.com/therecipe/qt/widgets"
)

// WebFrameHandler handle QWebFrame
type WebFrameHandler func(*webkit.QWebFrame)

// ResultHandler will pass the screenshot data to the callback func when finish screenshot
type ResultHandler func([]byte)

// ScreenshotConfig screenshot config
type ScreenshotConfig struct {
	URL              string
	Width            int
	Height           int
	Quality          int
	Format           string
	UA               string
	Timeout          time.Duration
	WebFrameHandlers []WebFrameHandler // handle QWebFrame when QWebPage Load finish
	ResultHandlers   []ResultHandler   // handle data when QWebPage scrennshot finish
}

// NewScreenshotConfig create a ScreenshotConfig
func NewScreenshotConfig(url string) *ScreenshotConfig {
	return &ScreenshotConfig{
		URL:     url,
		Width:   1920,
		Height:  1080,
		Quality: 50,
		Format:  "jpg",
		UA:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
		Timeout: 20 * time.Second,
	}
}

// WithWidth set width
func (c *ScreenshotConfig) WithWidth(width int) *ScreenshotConfig {
	c.Width = width
	return c
}

// WithHeight set height
func (c *ScreenshotConfig) WithHeight(height int) *ScreenshotConfig {
	c.Height = height
	return c
}

// WithQuality set quality
func (c *ScreenshotConfig) WithQuality(quality int) *ScreenshotConfig {
	c.Quality = quality
	return c
}

// WithFormat set format
func (c *ScreenshotConfig) WithFormat(foramt string) *ScreenshotConfig {
	c.Format = foramt
	return c
}

// WithUA set request useragent
func (c *ScreenshotConfig) WithUA(ua string) *ScreenshotConfig {
	c.UA = ua
	return c
}

// WithTimeout set request timeout
func (c *ScreenshotConfig) WithTimeout(timeout time.Duration) *ScreenshotConfig {
	c.Timeout = timeout
	return c
}

// RegisterWebFrameHandler register a WebFrameHandler to handlers
func (c *ScreenshotConfig) RegisterWebFrameHandler(funcs ...WebFrameHandler) {
	c.WebFrameHandlers = append(c.WebFrameHandlers, funcs...)
}

// RegisterResultHandler register a ResultHandler to handlers
func (c *ScreenshotConfig) RegisterResultHandler(funcs ...ResultHandler) {
	c.ResultHandlers = append(c.ResultHandlers, funcs...)
}

// FinishCallbackFunc will pass the screenshot data to the callback func when finish screenshot
type FinishCallbackFunc func([]byte)

// ScreenshotObject qt object
type ScreenshotObject struct {
	core.QObject

	_ func(config *ScreenshotConfig) `signal:"StartScreenshot"`
}

// NetworkAccessManager QNetworkAccessManager with timeout
// reference:
//   - phantomjs/src/networkaccessmanager.cpp:createRequest
//   - phantomjs/src/networkaccessmanager.cpp:handleTimeout
//   - https://www.cnblogs.com/apocelipes/p/9361690.html
//   - https://github.com/therecipe/examples/blob/master/advanced/widgets/tableview/main.go
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

	l.ConnectStartScreenshot(func(config *ScreenshotConfig) {
		l.GetScreenshot(config)
	})

	return l
}

// GetScreenshot get a snapshot for website
func (l *Loader) GetScreenshot(config *ScreenshotConfig) {
	url := config.URL
	width := config.Width
	height := config.Height
	imgQuality := config.Quality
	imgFormat := config.Format
	userAgent := config.UA
	timeout := config.Timeout
	webFrameHandlers := config.WebFrameHandlers
	resultHandlers := config.ResultHandlers
	page := webkit.NewQWebPage(nil)
	// indicate whether an error occurred that req is canceled
	isCancelReqError := false

	networkAccessManager := NewNetworkAccessManagerWithTimeout(page, timeout)
	networkAccessManager.registerErrorHandler(func(code network.QNetworkReply__NetworkError) {
		if code == network.QNetworkReply__OperationCanceledError {
			// when timeout error occur(cancel request)
			// 1. set the isCancelReqError indicate that req is canceled(timeout)
			// 2. call the ResultHandler with nil
			isCancelReqError = true
			for i := range resultHandlers {
				resultHandlers[i](nil)
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
		// handle the QWebFrame
		for i := range webFrameHandlers {
			webFrameHandlers[i](page.MainFrame())
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
		for i := range resultHandlers {
			resultHandlers[i](data)
		}
	})
}

// Screenshot get a website screenshot and return data
func (l *Loader) Screenshot(config *ScreenshotConfig) []byte {
	dataChan := make(chan []byte)
	// use the channel to synchronously get result
	once := new(sync.Once)
	fChanGetResult := ResultHandler(func(data []byte) {
		// avoid multiple calls that cause panic(send on closed chan)
		once.Do(func() {
			dataChan <- data
			close(dataChan)
		})
	})
	// insert the callback to head for getting result first and returnning
	config.ResultHandlers = append([]ResultHandler{fChanGetResult}, config.ResultHandlers...)
	l.StartScreenshot(config)
	screenshotBytes := <-dataChan
	fmt.Println(screenshotBytes)
	return screenshotBytes
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
