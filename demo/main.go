package main

import (
	"http_server"
	"http_server/request"
	"http_server/response"
	"time"
)

func main() {
	server := http_server.NewServer(":10001")

	go func() {
		server.Server(func(req *request.Request, res *response.Response) {
			res.HTML(200, "<h2>你好</h2>")
		})
	}()
	select {
	case <-time.After(time.Minute):
		server.Stop()
	}
}
