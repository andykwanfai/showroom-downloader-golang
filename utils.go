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
	"sort"
	"strconv"
	"strings"
	"time"
)

func handleError(err error) {
	timestamp := now().Format("2006-01-02 15:04:05")
	log.Println(timestamp, err)
}

func handleFatalError(err error) {
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

type httpClient func(string) (string, error)

func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		handleError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", errors.New(url + " Not Found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		handleError(err)
	}

	return string(body), nil
}

// get m3u8 from showroom website and api
func getM3u8Url(httpGet httpClient, showroomUrl string) (string, error) {
	html, _ := httpGet(showroomUrl)
	r, _ := regexp.Compile(`room_id=\d+`)
	roomId := r.FindString(html)

	if len(roomId) > 0 {
		var streamingUrls StreamingUrls
		resp, _ := httpGet(`https://www.showroom-live.com/api/live/streaming_url?room_id=` + roomId[8:])
		if resp == "{}" {
			fmt.Println("Waiting live stream start...")
			return "", nil
		}
		json.Unmarshal([]byte(resp), &streamingUrls)
		return streamingUrls.StreamingUrlList[0].Url, nil
	}
	return "", errors.New("cannot find room")
}

// get segment prefix from m3u8 url
func getUrlPrefix(m3u8Url string) string {
	index := strings.LastIndex(m3u8Url, "/")
	return m3u8Url[:index+1]
}

// download segment list from m3u8
func getSegmentList(httpGet httpClient, m3u8Url string) []string {
	playlist, _ := httpGet(m3u8Url)
	r, _ := regexp.Compile(`(?m)^\w*-\d+.ts`)
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
	r, _ := regexp.Compile(`\d+.ts`)

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

// get a list of all segment downloaded
func getAllSegments(successMap map[string]bool) []string {
	keys := make([]string, len(successMap))

	i := 0
	for k := range successMap {
		keys[i] = k
		i++
	}

	// sort the segments
	sort.Slice(keys, func(i, j int) bool {
		_, numA := getSegmentFormat(keys[i])
		_, numB := getSegmentFormat(keys[j])
		return numA < numB
	})

	return keys
}

func removeSegmentPrefix( segmentPrefix string) string {
	r, _ := regexp.Compile(`\d+.ts`)
	return r.FindString(segmentPrefix)
}
