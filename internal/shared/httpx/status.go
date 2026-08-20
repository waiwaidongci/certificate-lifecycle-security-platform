package httpx

import "net/http"

type StatusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Write(data []byte) (int, error) {
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	if r.status == http.StatusOK {
		r.status = http.StatusOK
	}
	return n, err
}

func (r *StatusRecorder) Status() int {
	return r.status
}

func (r *StatusRecorder) Bytes() int {
	return r.bytes
}
