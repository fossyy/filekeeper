package middleware

import (
	"bufio"
	"net"
	"net/http"
)

type HijackWriter struct {
	conn net.Conn
}

func NewHijackWriter(conn net.Conn) http.ResponseWriter {
	return &HijackWriter{conn: conn}
}

func (rw *HijackWriter) Header() http.Header {
	return http.Header{}
}

func (rw *HijackWriter) Write(bytes []byte) (int, error) {
	return rw.conn.Write(bytes)
}

func (rw *HijackWriter) WriteHeader(statusCode int) {
}

func (rw *HijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.conn, bufio.NewReadWriter(bufio.NewReader(rw.conn), bufio.NewWriter(rw.conn)), nil
}
