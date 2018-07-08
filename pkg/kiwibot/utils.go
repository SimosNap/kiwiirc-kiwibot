package kiwibot

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ShortenURL get a shorter url from git.io
func ShortenURL(targetURL string) string {
	resp, err := http.PostForm("https://git.io/", url.Values{"url": {targetURL}})
	if err != nil {
		fmt.Println(err)
		return targetURL
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return targetURL
	}

	var shortURL string
	if string(body) == targetURL {
		shortURL = resp.Header.Get("location")
	} else {
		shortURL = targetURL
	}
	return shortURL
}

// GetLast returns the last item from a slice
func GetLast(slice []string) string {
	return slice[len(slice)-1]
}

// ShortenSHA returns the last 7 chars from a sha string
func ShortenSHA(sha string) string {
	return sha[len(sha)-7:]
}

// Contains returns true/false if []string contains string
func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
