package response

import (
	"fmt"
	"http_server/request"
	"net"
	"net/http"
	"strings"
	"time"
)

type Header map[string]string

func (h Header) Del(key string) {
	delete(h, key)
}

func (h Header) Set(key string, value string) {
	h[key] = value
}

func (h Header) Has(key string) bool {
	if _, ok := h[key]; ok {
		return ok
	}
	return false
}

type Response struct {
	req  *request.Request
	conn *net.TCPConn
	Header
}

const contentType = "Content-Type"
const contentLength = "Content-Length"

func NewResponse(req *request.Request, conn *net.TCPConn) *Response {
	return &Response{
		req:    req,
		conn:   conn,
		Header: make(map[string]string),
	}
}
func (r *Response) JSON(status int, html string) {

}

func (r *Response) Request() *request.Request {
	return r.req
}

func (r *Response) HTML(status int, html string) {
	responseLine := r.getResponseLine(status)
	for key, value := range r.Header {
		key2 := strings.Title(key)
		if !r.Header.Has(key2) {
			r.Header.Del(key)
			r.Header[key2] = value
		}
	}
	if !r.Header.Has(contentType) {
		r.Header.Set(contentType, "text/html;charset=utf-8")
	}
	r.Header.Set(contentLength, fmt.Sprint(len([]byte(html))))
	var headers []string
	for key, value := range r.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, value))
	}
	lines := []string{responseLine}
	lines = append(lines, headers...)
	lines = append(lines, "", html)
	body := strings.Join(lines, "\r\n")
	r.writeString(body)
	r.printResponseLog(status)
}

func (r *Response) File(filePath string) {

}

// 响应行
func (r *Response) getResponseLine(status int) string {
	// 响应码文本
	statusText := http.StatusText(status)
	return strings.Join([]string{r.req.Proto, fmt.Sprint(status), statusText}, " ")
}
func (h Header) Get(key string, defaultValue string) string {
	if v, ok := h[key]; ok {
		return v
	}
	return defaultValue
}

func (r *Response) write(buf []byte) {
	_, _ = r.conn.Write(buf)
}
func (r *Response) writeString(msg string) {
	r.write([]byte(msg))
}
func (r *Response) printResponseLog(status int) {
	fmt.Printf(`%s - - [%s] "%s / %s" %d -%s`, r.req.Ip, time.Now().Format(time.RFC3339), r.req.Method, r.req.Proto, status, "\n")
}
