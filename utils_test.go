package main

import (
	"fmt"
	"reflect"
	"testing"
)

const M3u8Url = "https://hls-css.live.showroom-live.com/live/3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9.m3u8"

func httpGetM3u8Stub(url string) (string, error) {
	fmt.Println(url)
	if url == "streaming_url" {
		return "<html>room_id=12345678</html>", nil
	} else if url == "https://www.showroom-live.com/api/live/streaming_url?room_id=12345678" {
		return fmt.Sprintf("{\"streaming_url_list\":[{\"url\":\"%s\"}]}", M3u8Url), nil
	}

	if url == "not_streaming_url" {
		return "<html>room_id=11111111</html>", nil
	} else if url == "https://www.showroom-live.com/api/live/streaming_url?room_id=11111111" {
		return "{}", nil
	}

	return "", nil

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

	urlPrefix1 := getUrlPrefix("https://hls-css.live.showroom-live.com/live/3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9.m3u8")

	if urlPrefix1 != "https://hls-css.live.showroom-live.com/live/" {
		t.Fatal("invalid urlPrefix")
	}

	urlPrefix2 := getUrlPrefix("https://hls-origin230.showroom-cdn.com/liveedge/352abe11ba70fb8b73b79109bc3dcd48dea3107ef3f3cc924e27f95f6a3022de_source/chunklist.m3u8")

	if urlPrefix2 != "https://hls-origin230.showroom-cdn.com/liveedge/352abe11ba70fb8b73b79109bc3dcd48dea3107ef3f3cc924e27f95f6a3022de_source/" {
		t.Fatal("invalid urlPrefix")
	}

}

func httpGetPlaylistStub(url string) (string, error) {
	return `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:NO
#EXT-X-MEDIA-SEQUENCE:1723210480
#EXT-X-TARGETDURATION:3
#EXTINF:1.967,
3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210480.ts?txspiseq=106197697587716559759
#EXTINF:2.008,
3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210481.ts?txspiseq=106197697587716559759
#EXTINF:2.008,
3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210482.ts?txspiseq=106197697587716559759
`, nil
}

func TestGetSegmentList(t *testing.T) {

	expectedSegmentList := []string{
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210480.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210481.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210482.ts",
	}

	segmentList := getSegmentList(httpGetPlaylistStub, ".m3u8")

	if !reflect.DeepEqual(segmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}
}

func TestGetSegmentIndex(t *testing.T) {
	// segmentPrefix, segmentIndex := getSegmentFormat("media_b406154_5820.ts")

	// if segmentIndex != 5820 {
	// 	t.Fatal("invalid segment index")
	// }

	// if segmentPrefix != "3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210482" {
	// 	t.Fatal("invalid segment prefix " + segmentPrefix)
	// }

	segmentPrefix, segmentIndex := getSegmentFormat("3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1723210482.ts")

	if segmentIndex != 1723210482 {
		t.Fatal("invalid segment index")
	}

	if segmentPrefix != "3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-" {
		t.Fatal("invalid segment prefix")
	}
}

func TestGetOldSegmentList(t *testing.T) {
	oldSegmentList := getOldSegmentList("3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-", 5823, 3)

	expectedSegmentList := []string{
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5820.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5821.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5822.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}

	oldSegmentList = getOldSegmentList("3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-", 5823, 7)

	expectedSegmentList = []string{
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5816.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5817.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5818.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5819.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5820.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5821.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5822.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}

	oldSegmentList = getOldSegmentList("3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-", 3, 7)

	expectedSegmentList = []string{
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-1.ts",
		"3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-2.ts",
	}

	if !reflect.DeepEqual(oldSegmentList, expectedSegmentList) {
		t.Fatal("invalid segment list")
	}
}

func TestGetAllSegments(t *testing.T) {
	input := map[string]bool{
		"3cdabee90af8604-111.ts":  true,
		"3cdabee90af8604-3.ts":    true,
		"3cdabee90af8604-11.ts":   true,
		"3cdabee90af8604-6.ts":    true,
		"3cdabee90af8604-20.ts":   true,
		"3cdabee90af8604-10.ts":   true,
		"3cdabee90af8604-99.ts":   true,
		"3cdabee90af8604-1.ts":    true,
		"3cdabee90af8604-19.ts":   true,
		"3cdabee90af8604-100.ts":  true,
		"3cdabee90af8604-1000.ts": true,
		"3cdabee90af8604-21.ts":   true,
		"3cdabee90af8604-300.ts":  true,
		"3cdabee90af8604-32.ts":   true,
		"3cdabee90af8604-101.ts":  true,
		"3cdabee90af8604-110.ts":  true,
		"3cdabee90af8604-200.ts":  true,
	}
	expected := []string{
		"3cdabee90af8604-1.ts",
		"3cdabee90af8604-3.ts",
		"3cdabee90af8604-6.ts",
		"3cdabee90af8604-10.ts",
		"3cdabee90af8604-11.ts",
		"3cdabee90af8604-19.ts",
		"3cdabee90af8604-20.ts",
		"3cdabee90af8604-21.ts",
		"3cdabee90af8604-32.ts",
		"3cdabee90af8604-99.ts",
		"3cdabee90af8604-100.ts",
		"3cdabee90af8604-101.ts",
		"3cdabee90af8604-110.ts",
		"3cdabee90af8604-111.ts",
		"3cdabee90af8604-200.ts",
		"3cdabee90af8604-300.ts",
		"3cdabee90af8604-1000.ts",
	}
	result := getAllSegments(input)

	if !reflect.DeepEqual(result, expected) {
		t.Fatal("getAllSegments Error")
	}
}

func TestRemoveSegmentPrefix(t *testing.T) {
	input := "3cdabee90af8604be7e1045ba1ecc01cecd654429f97d487f4023d3025b863b9-5821.ts"
	expected := "5821.ts"
	result := removeSegmentPrefix(input)

	if result != expected {
		t.Fatal("removeSegmentPrefix Error")
	}
}