package httpx

import (
	"fmt"
	"io"
	"strings"
)

func responseBodyReader(body io.ReadCloser, limit int) (data []byte, err error) {
	if body == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	if limit <= 0 {
		limit = 64 * 1024
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()
	data, err = io.ReadAll(io.LimitReader(body, int64(limit)+1))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if len(data) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return []byte(strings.TrimSpace(string(data))), nil
}
