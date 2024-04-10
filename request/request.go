package request

import (
	"errors"
	"io"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

type Header map[string]string

func (h Header) Get(key string) string {
	if v, ok := h[key]; ok {
		return v
	}
	return ""
}
func (h Header) Values() []string {
	values := make([]string, len(h))
	for _, v := range h {
		values = append(values, v)
	}
	return values
}

type Request struct {
	Path          string            // 请求Path
	Proto         string            // HTTP协议  HTTP/1.1
	Method        string            // 请求方法
	Header        map[string]string // 请求头
	Body          io.Reader         // 请求体
	ContentType   string            // 内容类型
	ContentLength int64             // 内容长度
	Host          string            // Host
	RemoteAddr    string            // 远程客户端地址
	Ip            string            // 远程客户端IP
	form          url.Values
	body          []byte
	query         map[string]string // query参数
}

// 请求方法
var methods = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", "CONNECT", "TRACE", "OPTIONS"}

var notBodyMethods = []string{"GET", "HEAD", "OPTIONS"}

const sepChar = "\r\n"

// GetRequest 解析客户端发送过来的数据包
func GetRequest(conn *net.TCPConn, message string) (*Request, error) {
	// 第一部分请求行，第二部分请求头，第三部分请求体
	splits := strings.Split(message, sepChar)
	// 请求行
	requestLine := splits[0]
	splits2 := strings.Split(requestLine, " ")
	if len(splits2) != 3 {
		return nil, errors.New("请求行")
	}
	// 请求方法
	method := splits2[0]
	if !slices.Contains(methods, method) {
		return nil, errors.New("请求方法不存在")
	}
	// 请求地址
	path := splits2[1]
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("请求path错误")
	}
	// 请求协议
	proto := splits2[2]
	if !strings.HasPrefix(proto, "HTTP/") {
		return nil, errors.New("请求协议错误")
	}
	headers := make(map[string]string)
	var (
		bodyStr       string
		contentType   string
		contentLength int
	)
	for index, item := range splits[1:] {
		if item == "" {
			// 空行
			if index != len(splits) {
				// 请求体
				bodyStr = strings.Join(splits[index+1:], sepChar)
			}
			break
		} else {
			headerEntry := strings.Split(item, ": ")
			if len(headerEntry) != 2 {
				return nil, errors.New("请求头格式错误")
			}
			headerKey := strings.Title(headerEntry[0])
			value := headerEntry[1]
			if !slices.Contains(notBodyMethods, method) {
				// 有body才解析
				if v := strings.ReplaceAll(headerKey, "-", ""); v == "ContentType" {
					contentType = value
				}
				if v := strings.ReplaceAll(headerKey, "-", ""); v == "ContentLength" {
					contentLength, _ = strconv.Atoi(value)
				}
			}
			headers[headerKey] = value
		}
	}
	body := []byte(bodyStr)
	if contentLength == 0 && len(body) > 0 {
		// 没有在请求头找到ContentType
		contentLength = len(body)
	}
	remoteAddr := conn.RemoteAddr().String()
	ip := strings.Split(remoteAddr, ":")[0]
	req := &Request{
		Path:          path,
		Proto:         proto,
		Method:        method,
		RemoteAddr:    remoteAddr,
		Ip:            ip,
		Header:        headers,
		Body:          strings.NewReader(bodyStr),
		body:          body,
		form:          make(url.Values),
		query:         make(map[string]string),
		ContentType:   contentType,
		ContentLength: int64(contentLength),
	}
	// 解析body
	req.parseBody()
	// 解析query
	req.parseQuery()
	return req, nil
}

func (r *Request) GetBody() []byte {
	return r.body
}

func (r *Request) GetString() string {
	return string(r.body)
}

func (r *Request) parseQuery() {
	index := strings.Index(r.Path, "?")
	if index == -1 {
		return
	}
	queryStr := r.Path[index+1:]
	keyValueStrList := strings.Split(queryStr, "&")
	for _, keyValueStr := range keyValueStrList {
		s := strings.Split(keyValueStr, "=")
		if len(s) == 0 {
			continue
		}
		key := strings.TrimSpace(s[0])
		key = pathUnescape(key)

		if len(s) == 1 {
			r.query[key] = ""
			continue
		}
		r.query[key] = pathUnescape(strings.Join(s[1:], "="))
	}
}
func (r *Request) parseBody() {
	if r.ContentLength == 0 {
		return
	}
	body := r.GetString()
	if strings.HasPrefix(r.ContentType, "application/x-www-form-urlencoded") {
		// 解析到Form
		entryStrList := strings.Split(body, "&")
		for _, entryStr := range entryStrList {
			keyValue := strings.Split(entryStr, "=")
			if len(keyValue) == 0 {
				continue
			}
			if len(keyValue) == 1 {
				key := keyValue[0]
				r.form.Set(key, "")
				continue
			}
			key := strings.TrimSpace(keyValue[0])
			key = pathUnescape(key)
			value := strings.Join(keyValue[1:], "=")
			if strings.HasSuffix(key, "[]") {
				key = strings.TrimRight(key, "[]")
			}
			value = pathUnescape(value)
			if r.form.Has(key) {
				r.form.Add(key, value)
			} else {
				r.form.Set(key, value)
			}
		}
	}
}
func (r *Request) Query(key string) string {
	if v, ok := r.query[key]; ok {
		return v
	}
	return ""
}
func (r *Request) FormMap() url.Values {
	return r.form
}

func (r *Request) Form(key string) string {
	return r.form.Get(key)
}

func (r *Request) FormSlice(key string) []string {
	return r.form[key]
}

func (r *Request) QueryMap() map[string]string {
	return r.query
}

func pathUnescape(str string) string {
	if v, err := url.PathUnescape(str); err == nil {
		str = v
	}
	return str
}
