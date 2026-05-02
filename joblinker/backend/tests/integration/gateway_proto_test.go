package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"joblinker/internal/middleware"
	"joblinker/pkg/proto"
)

func TestGatewayProtobufContentNegotiation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		accept         string
		envMode        string
		wantProbuf     bool
	}{
		{
			name:       "protobuf accept when internal mode",
			accept:     "application/x-protobuf",
			envMode:    "internal",
			wantProbuf: true,
		},
		{
			name:       "json accept when internal mode",
			accept:     "application/json",
			envMode:    "internal",
			wantProbuf: false,
		},
		{
			name:       "protobuf accept when all mode",
			accept:     "application/x-protobuf",
			envMode:    "all",
			wantProbuf: true,
		},
		{
			name:       "protobuf accept when disabled",
			accept:     "application/x-protobuf",
			envMode:    "none",
			wantProbuf: false,
		},
		{
			name:       "empty accept defaults to json",
			accept:     "",
			envMode:    "internal",
			wantProbuf: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("PROTOBUF_ENABLED", tt.envMode)
			defer os.Unsetenv("PROTOBUF_ENABLED")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set("Accept", tt.accept)

			result := middleware.ShouldUseProtobuf(c)
			if result != tt.wantProbuf {
				t.Errorf("ShouldUseProtobuf() = %v, want %v", result, tt.wantProbuf)
			}
		})
	}
}

func TestGatewayParseAcceptHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		accept  string
		want    proto.ContentType
	}{
		{
			name:   "protobuf content type",
			accept: "application/x-protobuf",
			want:   proto.ContentTypeProtobuf,
		},
		{
			name:   "google protobuf",
			accept: "application/vnd.google.protobuf",
			want:   proto.ContentTypeProtobuf,
		},
		{
			name:   "json content type",
			accept: "application/json",
			want:   proto.ContentTypeJSON,
		},
		{
			name:   "empty accept",
			accept: "",
			want:   proto.ContentTypeJSON,
		},
		{
			name:   "protobuf with q value",
			accept: "application/x-protobuf;q=0.9",
			want:   proto.ContentTypeProtobuf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set("Accept", tt.accept)

			result := proto.ParseAcceptHeader(tt.accept)
			if result != tt.want {
				t.Errorf("ParseAcceptHeader(%q) = %v, want %v", tt.accept, result, tt.want)
			}
		})
	}
}

func TestEncodeProtobuf(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets protobuf content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := []byte{0x08, 0x01} // Protobuf varint
		err := middleware.EncodeProtobuf(c, data)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if w.Header().Get("Content-Type") != "application/x-protobuf" {
			t.Errorf("expected Content-Type 'application/x-protobuf', got %q", w.Header().Get("Content-Type"))
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestProtoMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("wraps response when protobuf enabled", func(t *testing.T) {
		os.Setenv("PROTOBUF_ENABLED", "internal")
		defer os.Unsetenv("PROTOBUF_ENABLED")

		r := gin.New()
		r.Use(middleware.GatewayMiddleware())
		r.Use(middleware.ProtoMiddleware())
		r.GET("/test", func(c *gin.Context) {
			if middleware.ShouldUseProtobuf(c) {
				_, ok := c.Get("protoWriter")
				if !ok {
					t.Error("protoWriter not set")
					return
				}
			}
			c.JSON(200, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept", "application/x-protobuf")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestProtobufRoundtrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("json response when no accept header", func(t *testing.T) {
		os.Setenv("PROTOBUF_ENABLED", "internal")
		defer os.Unsetenv("PROTOBUF_ENABLED")

		r := gin.New()
		r.Use(middleware.GatewayMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "hello"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No Accept header
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if !bytes.Contains([]byte(contentType), []byte("application/json")) {
			t.Errorf("expected JSON Content-Type, got %q", contentType)
		}

		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("response is not valid JSON: %v", err)
		}
		if resp["message"] != "hello" {
			t.Errorf("expected message 'hello', got %q", resp["message"])
		}
	})
}

func TestDetectContentTypeFromBody(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    proto.ContentType
	}{
		{
			name: "protobuf varint",
			data: []byte{0x08},
			want: proto.ContentTypeProtobuf,
		},
		{
			name: "json object",
			data: []byte(`{"key": "value"}`),
			want: proto.ContentTypeJSON,
		},
		{
			name: "empty",
			data: []byte{},
			want: proto.ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.DetectContentType(tt.data)
			if got != tt.want {
				t.Errorf("DetectContentType(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestReadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reads and restores body", func(t *testing.T) {
		r := gin.New()
		r.POST("/test", func(c *gin.Context) {
			body, err := middleware.ReadBody(c)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !bytes.Equal(body, []byte("test body")) {
				t.Errorf("expected 'test body', got %q", string(body))
			}
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("test body"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	})
}