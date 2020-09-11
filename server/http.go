package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Server ...
type Server struct {
	s *http.Server
}

// NewServer ...
func NewServer(listen string) *Server {

	s := &Server{}

	go func() {
		s.s = &http.Server{Addr: listen, Handler: s}
		s.s.ListenAndServe()
	}()
	return s
}

// Shutdown ...
func (s *Server) Shutdown() {
	zap.L().Debug("Shutting down Search Server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.s.Shutdown(ctx)
}

// Result ...
type Result struct {
	URLPath  string `json:"urlpath,omitempty"`
	Method   string `json:"method,omitempty"`
	ClientIP string `json:"clientip,omitempty"`
}

type contextKey struct {
	key string
}

// ConnContextKey ...
var ConnContextKey = &contextKey{"http-conn"}

// SaveConnInContext ...
func SaveConnInContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, ConnContextKey, c)
}

// GetConn ...
func GetConn(r *http.Request) net.Conn {
	return r.Context().Value(ConnContextKey).(net.Conn)
}

func (c *Result) String() string {
	pjson, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(pjson)
}

// ServeHTTP ...
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	result := &Result{}
	result.URLPath = req.URL.Path
	result.Method = req.Method
	result.ClientIP = getRemoteIP(req)

	conn := GetConn(req)

	zap.L().Debug(fmt.Sprintf("URLPath: %s, Method: %s, ClientIP: %s", result.URLPath, result.Method, result.ClientIP))

	fmt.Fprintf(w, result.String()+"\n")

}

func getRemoteIP(req *http.Request) string {
	forwarded := req.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return req.RemoteAddr
}
