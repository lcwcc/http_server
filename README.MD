## demo/main.go 测试
## 使用
```go
package main

import (
	"http_server"
	"http_server/request"
	"http_server/response"
	"time"
)

func main() {
	server := http_server.NewServer(":10001")
    server.Server(func(req *request.Request, res *response.Response) {
        res.HTML(200, "<h2>你好</h2>")
    })
}

```