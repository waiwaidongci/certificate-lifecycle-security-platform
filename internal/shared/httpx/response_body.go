package httpx

import "io"

func responseBodyReader(body io.ReadCloser, limit int) ([]byte, error) {
	return readAndCloseResponse(body, limit)
}
