package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	successMap := make(map[string]bool)

	url := os.Args[1]
	path := "./"
	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	dtString := now().Format("20060102")
	folderName := fmt.Sprintf("%s-showroom-%s", dtString, strings.Replace(url, "https://www.showroom-live.com/", "", 1))
	desFolder := fmt.Sprintf("%s%s/", path, folderName)

	m3u8Tick := time.Tick(10 * time.Second)

	var m3u8Url string

	for ; true; <-m3u8Tick {
		m3u8, err := getM3u8Url(httpGet, url)
		if err != nil {
			panic(err)
		}
		if m3u8 != "" {
			m3u8Url = m3u8
			break
		}
	}

	err := os.Mkdir(desFolder, 0755)
	if err != nil {
		handleError(err)
	}

	urlPrefix := getUrlPrefix(m3u8Url)

	downloadSegmentTick := time.Tick(2 * time.Second)
	oldSegmentRecoverTick := time.Tick(10 * time.Second)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for ; true; <-oldSegmentRecoverTick {
			oldSegmentRecover(m3u8Url, urlPrefix, desFolder, successMap)
		}
	}()

	go func() {
		for ; true; <-downloadSegmentTick {
			downloadNewSegments(m3u8Url, urlPrefix, desFolder, successMap)
		}
	}()

	<-sigs
	fmt.Println("exit")

	mergeSegments(successMap, desFolder, folderName)
}

func downloadSegments(
	m3u8Url string,
	urlPrefix string,
	desFolder string,
	successMap map[string]bool,
	segmentList []string) {

	for _, s := range segmentList {
		if successMap[s] == false {
			segment, err := httpGet(urlPrefix + s)
			if err != nil {
				handleError(err)
				return
			}
			writeFile(desFolder+s, []byte(segment))
			successMap[s] = true
		}
	}
}

func downloadNewSegments(m3u8Url string, urlPrefix string, desFolder string, successMap map[string]bool) {
	fmt.Println("download")
	segmentList := getSegmentList(httpGet, m3u8Url)
	downloadSegments(m3u8Url, urlPrefix, desFolder, successMap, segmentList)
}

// download last 50 segments
func oldSegmentRecover(m3u8Url string, urlPrefix string, desFolder string, successMap map[string]bool) {
	fmt.Println("oldSegmentRecover")
	pastSegmentLimit := 50

	segmentList := getSegmentList(httpGet, m3u8Url)

	if len(segmentList) == 0 {
		return
	}

	segmentPrefix, currentIndex := getSegmentFormat(segmentList[0])

	oldSegmentList := getOldSegmentList(segmentPrefix, currentIndex, pastSegmentLimit)

	downloadSegments(m3u8Url, urlPrefix, desFolder, successMap, oldSegmentList)
}

func mergeSegments(successMap map[string]bool, folder string, folderName string) {
	fmt.Println("merge segments")
	segments := getAllSegments(successMap)

	output, err := os.Create(filepath.Join(folder, folderName+".ts"))
	if err != nil {
		panic(err)
	}

	for _, segment := range segments {
		bytes, err := ioutil.ReadFile(filepath.Join(folder, segment))
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if _, err := output.Write(bytes); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
}
