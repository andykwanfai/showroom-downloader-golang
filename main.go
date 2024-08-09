package main

import (
	"fmt"
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
	urlSplits := strings.Split(url, "/")
	folderName := fmt.Sprintf("%s-showroom-%s", dtString, urlSplits[len(urlSplits)-1])
	desFolder := fmt.Sprintf("%s%s/", path, folderName)

	m3u8Tick := time.Tick(10 * time.Second)

	// wait the stream start
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

	// create the folder
	fmt.Println(desFolder)
	err := os.Mkdir(desFolder, 0755)
	if err != nil {
		handleFatalError(err)
	}

	urlPrefix := getUrlPrefix(m3u8Url)

	// create channel to wait for interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// download old segemnts
	oldSegmentRecoverTick := time.Tick(10 * time.Second)
	go func() {
		for ; true; <-oldSegmentRecoverTick {
			oldSegmentRecover(m3u8Url, urlPrefix, desFolder, successMap)
		}
	}()

	// download new segments
	downloadSegmentTick := time.Tick(2 * time.Second)
	go func() {
		for ; true; <-downloadSegmentTick {
			downloadNewSegments(m3u8Url, urlPrefix, desFolder, successMap)
		}
	}()

	// after interrupt signal
	<-sigs
	fmt.Println("exit")

	// merge downloaded segments
	mergeSegments(successMap, desFolder, folderName)
}

func downloadSegments(
	segmentPrefix string,
	urlPrefix string,
	desFolder string,
	successMap map[string]bool,
	segmentList []string) {

	for _, segmentName := range segmentList {
		fileName := removeSegmentPrefix(segmentName)
		if successMap[fileName] == false {
			segment, err := httpGet(urlPrefix + segmentName)
			if err != nil {
				handleError(err)
				continue
			}
			writeFile(desFolder+fileName, []byte(segment))
			successMap[fileName] = true
		}
	}
}

func downloadNewSegments(m3u8Url string, urlPrefix string, desFolder string, successMap map[string]bool) {
	fmt.Println("download")
	segmentList := getSegmentList(httpGet, m3u8Url)

	if len(segmentList) == 0 {
		return
	}

	segmentPrefix, _ := getSegmentFormat(segmentList[0])

	downloadSegments(segmentPrefix, urlPrefix, desFolder, successMap, segmentList)
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

	downloadSegments(segmentPrefix, urlPrefix, desFolder, successMap, oldSegmentList)
}

func mergeSegments(successMap map[string]bool, folder string, folderName string) {
	fmt.Println("merge segments")
	segments := getAllSegments(successMap)

	output, err := os.Create(filepath.Join(folder, folderName+".ts"))
	if err != nil {
		panic(err)
	}

	for _, segment := range segments {
		bytes, err := os.ReadFile(filepath.Join(folder, segment))
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
