package server

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// StatusBadRequest                   = 400 // RFC 7231, 6.5.1
// StatusUnauthorized                 = 401 // RFC 7235, 3.1
// StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
// StatusForbidden                    = 403 // RFC 7231, 6.5.3
// StatusNotFound                     = 404 // RFC 7231, 6.5.4
// StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
// StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
// StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
// StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
// StatusConflict                     = 409 // RFC 7231, 6.5.8
// StatusGone                         = 410 // RFC 7231, 6.5.9
// StatusLengthRequired               = 411 // RFC 7231, 6.5.10
// StatusPreconditionFailed           = 412 // RFC 7232, 4.2
// StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
// StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
// StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
// StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
// StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
// StatusTeapot                       = 418 // RFC 7168, 2.3.3
// StatusMisdirectedRequest           = 421 // RFC 7540, 9.1.2
// StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
// StatusLocked                       = 423 // RFC 4918, 11.3

// ServeHTTP HTTP/HTTPS Handler
func (t *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	zap.L().Debug(fmt.Sprintf("Entering ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	defer zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP path=%s method=%s", r.URL.Path, r.Method))

	w.Header().Set("Content-Type", "application/json")

	token := getBearerToken(r)

	if token == "" {
		http.Error(w, newErrorResponse("Token required")+"\n", http.StatusConflict)
		return
	}

	switch r.URL.Path {
	case "/getnonce":
		nonce, err := t.newNonce(r.Context(), token)
		if handleERR(w, err) {
			return
		}
		fmt.Fprintf(w, nonce.JSON()+"\n")
		return

	case "/getkeytab":

		principal := getKey(r, "principal")
		if principal == "" {
			http.Error(w, newErrorResponse("Principal required")+"\n", http.StatusConflict)
			return
		}

		keytab, err := t.getKeytab(r.Context(), token, principal)
		if handleERR(w, err) {
			return
		}

		fmt.Fprintf(w, keytab.JSON()+"\n")
		return
	}

	http.Error(w, newErrorResponse("Path "+r.URL.Path+" not mapped")+"\n", http.StatusConflict)

	zap.L().Debug(fmt.Sprintf("Exiting ServeHTTP"))
}

func newErrorResponse(message string) string {
	return "{\"error\":\"" + message + "\"}"
}

func handleERR(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	http.Error(w, newErrorResponse(err.Error())+"\n", http.StatusConflict)
	return true
}

func getBearerToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if token != "" {
		return token
	}
	return getKey(r, "bearertoken")
}

func getKey(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return string(keys[0])
}