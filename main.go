package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/corona10/goimagehash"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"gorm.io/gorm"

	"github.com/therecipe/qt/webkit"

	"github.com/akkuman/webkit-screenshot/wk"
)

// DB gorm DB instance
var DB *gorm.DB

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	pflag.Int("thread", 30, "thread number")
	pflag.String("file", "", "urls file which will be execute website screenshot")
	pflag.String("taskid", "", "id of task")
	pflag.String("db.dialect", "sqlite3", "database type")
	pflag.String("db.dsn", "res.db", "database dsn")
	viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}
	pflag.Parse()
}

func main() {
	var err error
	DB, err = InitDb()
	if err != nil {
		panic(err)
	}
	loader := wk.NewLoader()

	urls, err := getURLs(viper.GetString("file"))
	if err != nil {
		panic(err)
	}
	amount := len(urls)
	threadNum := viper.GetInt("thread")
	taskid := viper.GetString("taskid")
	fmt.Println(threadNum)
	if threadNum <= 0 {
		threadNum = amount
	}
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < int(math.Ceil(float64(amount)/float64(threadNum))); i++ {
			start := i * threadNum
			end := (i + 1) * threadNum
			if end > amount {
				end = amount
			}
			wg.Add(end - start)
			for _, url := range urls[start:end] {
				screenshotItem := new(Screenshot)
				screenshotItem.TaskID = taskid
				screenshotItem.URL = url
				screenshotItem.URLXxhash = int64(getXXHash(url))
				screenshotItem.TaskIDXxhash = int64(getXXHash(taskid))

				config := wk.NewScreenshotConfig(url).WithTimeout(30 * time.Second)
				config.RegisterWebFrameHandler(
					wk.WebFrameHandler(waitSingleApp),
					wk.WebFrameHandler(func(frame *webkit.QWebFrame) {
						screenshotItem.HTML = frame.ToHtml()
						screenshotItem.PlainText = frame.ToPlainText()
					}),
				)

				go getScreenshot(loader, config, wg, screenshotItem)
			}
			wg.Wait()
		}
		loader.Exit(0)
	}()

	loader.Exec()
}

func waitSingleApp(frame *webkit.QWebFrame) {
	htmlContent := frame.ToHtml()
	if strings.Contains(htmlContent, "doesn't work properly without JavaScript enabled. Please enable it to continue.") {
		time.Sleep(2 * time.Second)
	}
}

func getURLs(fpath string) (urls []string, err error) {
	var fi *os.File
	fi, err = os.Open(fpath)
	if err != nil {
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		urls = append(urls, string(a))
	}
	return
}

func getPHash(data []byte) uint64 {
	img, _ := jpeg.Decode(bytes.NewReader(data))
	h, _ := goimagehash.PerceptionHash(img)
	return h.GetHash()
}

func getXXHash(data string) uint64 {
	h := xxhash.New64()
	r := strings.NewReader(data)
	io.Copy(h, r)
	return h.Sum64()
}

func getScreenshot(loader *wk.Loader, config *wk.ScreenshotConfig, wg *sync.WaitGroup, s *Screenshot) {
	data := loader.Screenshot(config)
	if data != nil {
		s.Phash = int64(getPHash(data))
		filename := fmt.Sprintf("%s_%d.jpg", time.Now().Format("20060102150405"), uint64(s.URLXxhash))
		fpath, _ := savefile(filename, data)
		s.ImgPath = fpath
		fmt.Println(config.URL, "done!!!")
		DB.Create(s)
	}
	wg.Done()
}

func savefile(filename string, data []byte) (fpath string, err error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	fpath = fmt.Sprintf("./screenshots/%d/%d/%s", year, month, filename)
	dir := filepath.Dir(fpath)
	_, err = os.Lstat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	err = ioutil.WriteFile(fpath, data, 0644)
	return
}
