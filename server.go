package http_server

import (
	"fmt"
	"http_server/request"
	"http_server/response"
	"io"
	"log/slog"
	"net"
)

type Server struct {
	Addr string
	HandleFunc
	listen *net.TCPListener
	close  chan struct{}
}

type HandleFunc func(request *request.Request, response *response.Response)

// NewServer 创建一个服务
func NewServer(addr string) *Server {
	return &Server{
		Addr:  addr,
		close: make(chan struct{}, 1),
	}
}

// Server 启动服务
func (s *Server) Server(handleFunc HandleFunc) {
	if handleFunc == nil {
		panic("handleFunc not nil")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", s.Addr)
	listen, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		panic(err)
	}
	s.listen = listen
	s.HandleFunc = handleFunc
	slog.Info(fmt.Sprintf("Server Start... Listen %s", listen.Addr().String()))
	go s.start()
	select {
	case <-s.close:
		s.listen.Close()
	}
}

// Stop 停止服务
func (s *Server) Stop() {
	s.close <- struct{}{}
}

// 启动服务 接收每个TCP链接 开启协程处理请求
func (s *Server) start() {
	for {
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			slog.Error(fmt.Sprint(err))
			continue
		}
		go s.readBuf(conn)
	}
}
func (s *Server) readBuf(conn *net.TCPConn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			slog.Error("Client Disconnect")
			return
		}
		slog.Error("Read " + fmt.Sprint(err))
		return
	}
	req, err := request.GetRequest(conn, string(buf[:n]))
	if err != nil {
		slog.Error(fmt.Sprintf("数据包解析失败，并不是HTTP请求!!!，Error：%s", err))
		return
	}
	res := response.NewResponse(req, conn)
	s.HandleFunc(req, res)
	return
}
