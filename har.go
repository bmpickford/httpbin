package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var startingEntrySize int = 1000

type Har struct {
	HarLog HarLog `json:"log"`
}

func (har *Har) format() string {
	formattedHar, err := json.MarshalIndent(har, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(formattedHar)
}

type HarLog struct {
	Version string     `json:"version"`
	Creator string     `json:"creator"`
	Browser string     `json:"browser"`
	Pages   []HarPage  `json:"pages"`
	Entries []HarEntry `json:"entries"`
}

func newHarLog() *HarLog {
	harLog := HarLog{
		Version: "1.2",
		Creator: "HTTPBin",
		Browser: "",
		Pages:   make([]HarPage, 0, 10),
		Entries: makeNewEntries(),
	}
	return &harLog
}

func (harLog *HarLog) addEntry(entry ...HarEntry) {
	entries := harLog.Entries
	m := len(entries)
	n := m + len(entry)
	if n > cap(entries) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newEntries := make([]HarEntry, (n+1)*2)
		copy(newEntries, entries)
		entries = newEntries
	}
	entries = entries[0:n]
	copy(entries[m:n], entry)
	harLog.Entries = entries
	fmt.Println("Request received ", entry[0].Request.Method, entry[0].Request.Url)
}

func makeNewEntries() []HarEntry {
	return make([]HarEntry, 0, startingEntrySize)
}

type HarPage struct {
	Id              string         `json:"id"`
	StartedDateTime time.Time      `json:"startedDateTime"`
	Title           string         `json:"title"`
	PageTimings     HarPageTimings `json:"pageTimings"`
}

type HarEntry struct {
	PageRef         string       `json:"pageRef"`
	StartedDateTime time.Time    `json:"startedDateTime"`
	Time            int64        `json:"time"`
	Request         *HarRequest  `json:"request"`
	Response        *HarResponse `json:"response"`
	Timings         HarTimings   `json:"timings"`
	ServerIpAddress string       `json:"serverIpAddress"`
	Connection      string       `json:"connection"`
}

type HarRequest struct {
	Method      string             `json:"method"`
	Url         string             `json:"url"`
	HttpVersion string             `json:"httpVersion"`
	Cookies     []HarCookie        `json:"cookies"`
	Headers     []HarNameValuePair `json:"headers"`
	QueryString []HarNameValuePair `json:"queryString"`
	PostData    *HarPostData       `json:"postData"`
	BodySize    int64              `json:"bodySize"`
	HeadersSize int64              `json:"headersSize"`
}

type HarResponse struct {
	Status      int                `json:"status"`
	StatusText  string             `json:"statusText"`
	HttpVersion string             `json:"httpVersion"`
	Cookies     []HarCookie        `json:"cookies"`
	Headers     []HarNameValuePair `json:"headers"`
	Content     *HarContent        `json:"content"`
	RedirectUrl string             `json:"redirectUrl"`
	BodySize    int64              `json:"bodySize"`
	HeadersSize int64              `json:"headersSize"`
}

type HarCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path"`
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires"`
	HttpOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}

type HarNameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HarPostData struct {
	MimeType string             `json:"mimeType"`
	Params   []HarPostDataParam `json:"params"`
	Text     string             `json:"text"`
}

type HarPostDataParam struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type HarContent struct {
	Size        int64  `json:"size"`
	Compression int64  `json:"compression"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text"`
	Encoding    string `json:"encoding"`
}

type HarPageTimings struct {
	OnContentLoad int64 `json:"onContentLoad"`
	OnLoad        int64 `json:"onLoad"`
}

type HarTimings struct {
	Blocked int64
	Dns     int64
	Connect int64
	Send    int64
	Wait    int64
	Receive int64
	Ssl     int64
}

func parseEntry(req *http.Request) *HarEntry {
	harEntry := new(HarEntry)
	harEntry.Request = parseRequest(req)
	harEntry.StartedDateTime = time.Now()
	harEntry.Response = nil
	harEntry.Time = time.Now().Unix()

	return harEntry
}

func parseRequest(req *http.Request) *HarRequest {
	if req == nil {
		return nil
	}
	harRequest := HarRequest{
		Method:      req.Method,
		Url:         req.URL.String(),
		HttpVersion: req.Proto,
		Cookies:     parseCookies(req.Cookies()),
		Headers:     parseStringArrMap(req.Header),
		QueryString: parseStringArrMap((req.URL.Query())),
		BodySize:    req.ContentLength,
		HeadersSize: calcHeaderSize(req.Header),
	}

	if req.Method == "POST" || req.Method == "PUT" {
		harRequest.PostData = parsePostData(req)
	}

	return &harRequest
}

func calcHeaderSize(header http.Header) int64 {
	headerSize := 0
	for headerName, headerValues := range header {
		headerSize += len(headerName) + 2
		for _, v := range headerValues {
			headerSize += len(v)
		}
	}
	return int64(headerSize)
}

func parsePostData(req *http.Request) *HarPostData {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("Error parsing request to %v: %v\n", req.URL, e)
		}
	}()

	harPostData := new(HarPostData)
	contentType := req.Header["Content-Type"]
	if contentType == nil {
		panic("Missing content type in request")
	}
	harPostData.MimeType = contentType[0]

	if len(req.PostForm) > 0 {
		index := 0
		params := make([]HarPostDataParam, len(req.PostForm))
		for k, v := range req.PostForm {
			param := HarPostDataParam{
				Name:  k,
				Value: strings.Join(v, ","),
			}
			params[index] = param
			index++
		}
		harPostData.Params = params
	} else {
		str, _ := ioutil.ReadAll(req.Body)
		harPostData.Text = string(str)
	}
	return harPostData
}

func parseStringArrMap(stringArrMap map[string][]string) []HarNameValuePair {
	index := 0
	harQueryString := make([]HarNameValuePair, len(stringArrMap))
	for k, v := range stringArrMap {
		escapedKey, _ := url.QueryUnescape(k)
		escapedValues, _ := url.QueryUnescape(strings.Join(v, ","))
		harNameValuePair := HarNameValuePair{
			Name:  escapedKey,
			Value: escapedValues,
		}
		harQueryString[index] = harNameValuePair
		index++
	}
	return harQueryString
}

func parseCookies(cookies []*http.Cookie) []HarCookie {
	harCookies := make([]HarCookie, len(cookies))
	for i, cookie := range cookies {
		harCookie := HarCookie{
			Name:     cookie.Name,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires,
			HttpOnly: cookie.HttpOnly,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			Value:    cookie.Value,
		}
		harCookies[i] = harCookie
	}
	return harCookies
}
