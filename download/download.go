package download

import (
	"../converter"
	"../utils"
	"errors"
	"fmt"
	"github.com/cnych/starjazz/mathx"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UrlSave(vfile, url string, header http.Header) (result string, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("get video info error: ", err) // 这里的err其实就是panic传入的内容
		}
		//resp.Body.close()
	}()
	//log.Println("downloading ", vfile)
	for i := 0; i < 3; i++ {
		_, resp := utils.RequestUrl(url, header)
		contentLength, _ := strconv.ParseInt(resp.Header["Content-Length"][0], 10, 64)
		f, _ := os.Create(vfile)
		io.Copy(f, resp.Body)
		if fileInfo, err := os.Stat(vfile); err == nil {
			fileLength := fileInfo.Size()
			if contentLength == fileLength {
				result = vfile
				break
			} else {
				//log.Println("file size not equal")
				log.Printf("file size not equal %s,%d,%s ", vfile, fileLength, contentLength)
			}
		} else {
			log.Println("vfile not exists ", vfile)
		}
	}
	if result != vfile {
		err = errors.New("download video error")
	}
	return
}

func DownloadUrls(urls []string, ext string, info map[string]interface{}) (vfile string, err error) {
	title := info["title"].(string)
	vfile = "output" + "." + ext
	var header http.Header
	if h, ok := info["header"]; ok {
		header = h.(http.Header)
	}
	if len(urls) == 1 {
		vfile, err = UrlSave(vfile, urls[0], header)
	} else {
		var vfiles []string
		for index, url := range urls {
			f := fmt.Sprintf("%s_%d.%s", "output", index, ext)
			vf, err := UrlSave(f, url, header)
			if err == nil {
				vfiles = append(vfiles, vf)
			} else {
				err = errors.New(fmt.Sprintf("download %s fail", f))
			}
		}
		if len(vfiles) == len(urls) {
			options := map[string]interface{}{"format": "mp4"}
			audio := map[string]string{"codec": "copy"}
			options["audio"] = audio
			video := map[string]string{"codec": "copy", "faststart": "true"}
			options["video"] = video
			conv := converter.FFMpeg{}
			result := conv.Merge(vfiles, vfile, options)
			if !result {
				err = errors.New("Merge videos error")
			}
			for _, v := range vfiles {
				os.Remove(v)
			}
		}
	}
	os.Rename(vfile, fmt.Sprintf("%s.%s", title, ext))
	return
}

func UrlSize(urls []string, header http.Header) (size int64) {
	for _, url := range urls {
		_, resp := utils.RequestUrl(url, header)
		contentLength, _ := strconv.ParseInt(resp.Header["Content-Length"][0], 10, 64)
		size += contentLength
	}
	return
}

func Download(urls []string, ext string, info map[string]interface{}) error {
	title := info["title"].(string)
	fmt.Printf("\n")
	fmt.Printf("site:				%v\n", info["site"])
	fmt.Printf("title:				%s\n", title)
	fmt.Printf("type:				%v\n", info["type"])
	fmt.Printf("urls:				%v\n", len(urls))
	var header http.Header
	if h, ok := info["header"]; ok {
		header = h.(http.Header)
	}
	size := UrlSize(urls, header)
	s := fmt.Sprintf("%.2f MiB (%d bytes)", mathx.Round(float64(size)/1024/1024, 2), size)
	fmt.Printf("size:				%v\n", s)
	fmt.Printf("Downloading %s ...\n", title)
	start := time.Now().Unix()
	DownloadUrls(urls, ext, info)
	end := time.Now().Unix()
	cost := end - start
	fmt.Printf("Saving Me at the %s ...Done.Cost %vs\n", title, cost)
	fmt.Printf("\n")
	return nil
}
