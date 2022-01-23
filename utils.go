package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func handleError(err error) {
	timestamp := now().Format("2006-01-02 15:04:05")
	log.Fatalln(timestamp, err)
}

func now() time.Time {
	return time.Now()
}

type Url struct {
	Url string `json:"url"`
}

type StreamingUrls struct {
	StreamingUrlList []Url `json:"streaming_url_list"`
}

type httpClient func(string) string

func httpGet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		handleError(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		handleError(err)
	}

	return string(body)
}

// get m3u8 from showroom website and api
func getM3u8Url(httpGet httpClient, showroomUrl string) (string, error) {
	html := httpGet(showroomUrl)
	r, _ := regexp.Compile("room_id=\\d+")
	roomId := r.FindString(html)

	if len(roomId) > 0 {
		var streamingUrls StreamingUrls
		resp := httpGet(`https://www.showroom-live.com/api/live/streaming_url?room_id=` + roomId[8:])
		if resp == "{}" {
			fmt.Println("Waiting live stream start...")
			return "", nil
		}
		json.Unmarshal([]byte(resp), &streamingUrls)
		return streamingUrls.StreamingUrlList[0].Url, nil
	}
	return "", errors.New("Cannot find room")
}

// get segment prefix from m3u8 url
func getUrlPrefix(m3u8Url string) string {
	r, _ := regexp.Compile("chunklist\\w*.m3u8")
	return r.ReplaceAllString(m3u8Url, "")
}

// download segment list from m3u8
func getSegmentList(httpGet httpClient, m3u8Url string) []string {
	playlist := httpGet(m3u8Url)
	r, _ := regexp.Compile("media_\\w+.ts")
	segmentList := r.FindAllString(playlist, -1)
	return segmentList
}

func writeFile(name string, data []byte) {
	fmt.Println(name)
	err := os.WriteFile(name, data, 0777)
	if err != nil {
		handleError(err)
	}
}

// get prefix and index of segment from a segment name
func getSegmentFormat(segment string) (prefix string, index int) {
	r, _ := regexp.Compile("\\d+.ts")

	indexTs := r.FindString(segment)

	prefix = strings.ReplaceAll(segment, indexTs, "")

	indexStr := strings.ReplaceAll(indexTs, ".ts", "")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		index = 1
	}

	return prefix, index
}

// get old segment list by current segment index
func getOldSegmentList(segmentPrefix string, currentSegmentIndex int, pastSegmentLimit int) []string {

	start := currentSegmentIndex - pastSegmentLimit

	if start < 1 {
		start = 1
	}

	var oldSegmentList []string

	for i := start; i < currentSegmentIndex; i++ {
		oldSegmentList = append(oldSegmentList, fmt.Sprintf("%s%d.ts", segmentPrefix, i))
	}

	return oldSegmentList
}
