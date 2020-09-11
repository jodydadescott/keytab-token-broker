package httpserver

import (
	"context"
	"fmt"
	"kbridge/internal/api"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPServer ...
type HTTPServer struct {
	httpServer *http.Server
	api        *api.API
}

// JSON ...
type JSON interface {
	JSON() string
}

// NewHTTPServer ...
func NewHTTPServer(listen string, api *api.API) *HTTPServer {

	zap.L().Debug("Starting")

	t := &HTTPServer{
		api: api,
	}

	go func() {
		t.httpServer = &http.Server{Addr: listen, Handler: t}
		t.httpServer.ListenAndServe()
	}()
	return t
}

// Shutdown ...
func (t *HTTPServer) Shutdown() {
	zap.L().Debug("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.httpServer.Shutdown(ctx)
}

func newErrorResponse(message string) string {
	return "{\"error\":\"" + message + "\"}"
}

type x struct {
	w     http.ResponseWriter
	r     *http.Request
	token string
	api   *api.API
}

func (t *x) getBearerToken() (string, error) {

	if t.token != "" {
		return t.token, nil
	}

	token := t.r.Header.Get("Authorization")
	if token != "" {
		t.token = token
		return t.token, nil
	}

	token = t.getKey("bearertoken")
	if token != "" {
		t.token = token
		return t.token, nil
	}

	return "", fmt.Errorf("Token not present")
}

// func (t *x) getNonce() (string, error) {

// 	if t.nonce != "" {
// 		return t.nonce, nil
// 	}

// 	nonce := t.getKey("nonce")
// 	if nonce != "" {
// 		t.nonce = nonce
// 		return t.nonce, nil
// 	}

// 	return "", fmt.Errorf("Nonce not found")
// }

// ServeHTTP ...
func (t *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	zap.L().Debug(fmt.Sprintf("Entering ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	w.Header().Set("Content-Type", "application/json")

	x := &x{
		w:   w,
		r:   r,
		api: t.api,
	}

	var err error
	var json JSON

	switch r.URL.Path {
	case "/getnonce":
		json, err = x.getnonce()

	case "/getkeytab":
		json, err = x.getkeytab()

	default:
		fmt.Fprintf(w, newErrorResponse("Path "+r.URL.Path+" not mapped")+"\n")
		w.WriteHeader(http.StatusConflict)
		return
	}

	if err != nil {
		fmt.Fprintf(w, newErrorResponse(err.Error())+"\n")
		w.WriteHeader(http.StatusConflict)
		return
	}

	if json == nil {
		panic("this should not happen")
	}

	fmt.Fprintf(w, json.JSON()+"\n")
	w.WriteHeader(http.StatusOK)

	zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP"))

}

func (t *x) getnonce() (JSON, error) {

	token, err := t.getBearerToken()
	if err != nil {
		return nil, err
	}

	nonce, err := t.api.NewNonce(token)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func (t *x) getkeytab() (JSON, error) {

	token, err := t.getBearerToken()
	if err != nil {
		return nil, err
	}

	principal := t.getKey("principal")
	if principal == "" {
		return nil, fmt.Errorf("Principal not found")
	}

	keytab, err := t.api.GetKeytab(token, principal)
	if err != nil {
		return nil, err
	}
	return keytab, nil
}

func (t *x) getKey(name string) string {

	keys, ok := t.r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return string(keys[0])
}
