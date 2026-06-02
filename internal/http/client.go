package httpc

import (
	"net/http"
	"time"
)

var HTTPClient = &http.Client{Timeout: 30 * time.Second}
