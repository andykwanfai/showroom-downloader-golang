package main

import (
	"fmt"
	"reflect"
	"testing"
)

const M3u8Url = "https://hls-origin203.showroom-cdn.com/liveedge/test/chunklist.m3u8"

func httpGetM3u8Stub(url string) string {
	fmt.Println(url)
	if url == "streaming_url" {
		return "<html>room_id=12345678</html>"
	} else if url == "https://www.showroom-live.com/api/live/streaming_url?room_id=12345678" {
		return "{\"streaming_url_list\":[{\"url\":\"https://hls-origin203.showroom-cdn.com/liveedge/test/chunklist.m3u8\"}]}"
	}

	if url == "not_streaming_url" {
		return "<html>room_id=11111111</html>"
	} else if url == "https://www.showroom-live.com/api/live/streaming_url?room_id=11111111" {
		return "{}"
	}

	return ""

}

func TestGetM3u8(t *testing.T) {

	m3u8_url, err := getM3u8Url(httpGetM3u8Stub, "streaming_url")

	if m3u8_url != M3u8Url {
		t.Fatal("invalid m3u8")
	}

	m3u8_url, err = getM3u8Url(httpGetM3u8Stub, "invalid_showroom_url")
	if err == nil {
		t.Fatal("Invalid showroom url should return error")
	}

	m3u8_url, err = getM3u8Url(httpGetM3u8Stub, "not_streaming_url")
	if m3u8_url != "" {
		t.Fatal("not_streaming_url should return empty")
	}
}

func TestGetUrlPrefix(t *testing.T) {

	urlPrefix1 := getUrlPrefix("https://hls-origin250.showroom-cdn.com/liveedge/ngrp:1d1698c4f1670ff4a90a1315183117636a6ee4b1387d645a49aec3fd8ca9465e_all/chunklist_b341269.m3u8")

	if urlPrefix1 != "https://hls-origin250.showroom-cdn.com/liveedge/ngrp:1d1698c4f1670ff4a90a1315183117636a6ee4b1387d645a49aec3fd8ca9465e_all/" {
		t.Fatal("invalid urlPrefix")
	}

	urlPrefix2 := getUrlPrefix("https://hls-origin230.showroom-cdn.com/liveedge/352abe11ba70fb8b73b79109bc3dcd48dea3107ef3f3cc924e27f95f6a3022de_source/chunklist.m3u8")

	if urlPrefix2 != "https://hls-origin230.showroom-cdn.com/liveedge/352abe11ba70fb8b73b79109bc3dcd48dea3107ef3f3cc924e27f95f6a3022de_source/" {
		t.Fatal("invalid urlPrefix")
	}

}

func httpGetPlaylistStub(url string) string {
	return `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:4
#EXT-X-MEDIA-SEQUENCE:5820
#EXT-X-PROGRAM-DATE-TIME:2021-11-10T01:14:17.078+09:00
#EXTINF:2.0,
media_b406154_5820.ts
#EXTINF:2.0,
media_b406154_5821.ts
#EXTINF:2.0,
media_b406154_5822.ts
#EXTINF:2.0,
media_b406154_5823.ts
#EXTINF:2.0,
media_b406154_5824.ts
`
}

func TestGetSegmentList(t *testing.T) {

	expectedSegmentList := []string{
		"media_b406154_5820.ts",
		"media_b406154_5821.ts",
		"media_b406154_5822.ts",
		"media_b406154_5823.ts",
		"media_b406154_5824.ts",
	}

	segmentList := getSegmentList(httpGetPlaylistStub, ".m3u8")

	if !reflect.DeepEqual(segmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}
}

func TestGetSegmentIndex(t *testing.T) {
	segmentPrefix, segmentIndex := getSegmentFormat("media_b406154_5820.ts")

	if segmentIndex != 5820 {
		t.Fatal("invalid segment index")
	}

	if segmentPrefix != "media_b406154_" {
		t.Fatal("invalid segment prefix " + segmentPrefix)
	}

	segmentPrefix, segmentIndex = getSegmentFormat("media_5820.ts")

	if segmentIndex != 5820 {
		t.Fatal("invalid segment index")
	}

	if segmentPrefix != "media_" {
		t.Fatal("invalid segment prefix")
	}
}

func TestGetOldSegmentList(t *testing.T) {
	oldSegmentList := getOldSegmentList("media_b406154_", 5823, 3)

	expectedSegmentList := []string{
		"media_b406154_5820.ts",
		"media_b406154_5821.ts",
		"media_b406154_5822.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}

	oldSegmentList = getOldSegmentList("media_b406154_", 5823, 7)

	expectedSegmentList = []string{
		"media_b406154_5816.ts",
		"media_b406154_5817.ts",
		"media_b406154_5818.ts",
		"media_b406154_5819.ts",
		"media_b406154_5820.ts",
		"media_b406154_5821.ts",
		"media_b406154_5822.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}

	oldSegmentList = getOldSegmentList("media_b406154_", 3, 7)

	expectedSegmentList = []string{
		"media_b406154_1.ts",
		"media_b406154_2.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}
}
