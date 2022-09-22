package ehttp

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// 执行请求的http客户端对象
var theHttpClient *http.Client

// 请求的时候需要携带的Cookie
var theCookieJar *cookiejar.Jar

// 伪造的 User-Agent，假装是一个浏览器
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36 Edg/103.0.1264.62"

var customHeaders map[string]string

var encoding = "utf-8"

// 初始化Cookie对象和http客户端
func init() {
	theCookieJar, _ = cookiejar.New(nil)
	theHttpClient = &http.Client{
		Jar: theCookieJar,
	}
}

// SetEncoding 设置编码
func SetEncoding(encode string) {
	encoding = encode
}

// SetCookies 设置Cookie
func SetCookies(u *url.URL, cookies []*http.Cookie) {
	theCookieJar.SetCookies(u, cookies)
}

func ClearCustomHeaders() {
	customHeaders = make(map[string]string, 0)
}

func SetCustomHeaders(hmap map[string]string) {
	customHeaders = hmap
}

// Get 使用GET方式发起HTTP请求
func Get(location string) (string, error) {
	return RetryGet(location, 1)
}

// Post 使用GET方式发起HTTP请求
func Post(location string, body string) (string, error) {
	return RetryPost(location, []byte(body), 1)
}

// RetryGet 使用GET方式发起HTTP请求，如果请求失败则按照给定的重试次数执行重发，如果已经达到最大重试次数则返回错误
func RetryGet(location string, retry int) (string, error) {
	req, err := http.NewRequest("GET", location, nil)
	if err == nil {
		return execRequest(req, nil, retry)
	} else {
		return "", err
	}
}

// RetryPost Post请求获取HTML内容
func RetryPost(location string, body []byte, retry int) (string, error) {
	req, err := http.NewRequest("POST", location, bytes.NewBuffer(body))
	if err == nil {
		return execRequest(req, body, retry)
	} else {
		return "", err
	}
}

// 执行HTTP请求
func execRequest(req *http.Request, body []byte, retry int) (string, error) {
	req.Header.Set("User-Agent", userAgent)
	if customHeaders != nil && len(customHeaders) > 0 {
		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}
	}
	if body != nil && len(body) > 0 {
		req.Header.Set("Content-Type", http.DetectContentType(body))
	}
	var retErr error
	resp, err := theHttpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		var body []byte
		var err error
		if strings.Contains(strings.ToLower(encoding), "gb") {
			utf8Reader := transform.NewReader(resp.Body,
				simplifiedchinese.GBK.NewDecoder())
			body, err = ioutil.ReadAll(utf8Reader)
		} else {
			body, err = ioutil.ReadAll(resp.Body)
		}
		if err == nil {
			return string(body), nil
		} else {
			retErr = err
		}
	} else {
		retErr = err
	}
	retry--
	if retry > 0 {
		return execRequest(req, body, retry)
	} else {
		return "", retErr
	}
}
