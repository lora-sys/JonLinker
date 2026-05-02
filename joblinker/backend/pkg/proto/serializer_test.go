package proto

import (
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want ContentType
	}{
		{
			name: "empty data returns json",
			data: []byte{},
			want: ContentTypeJSON,
		},
		{
			name: "json object",
			data: []byte(`{"key": "value"}`),
			want: ContentTypeJSON,
		},
		{
			name: "protobuf varint field 1 wire type 0",
			data: []byte{0x08},
			want: ContentTypeProtobuf,
		},
		{
			name: "protobuf varint field 1 wire type 0 second byte",
			data: []byte{0x0F},
			want: ContentTypeProtobuf,
		},
		{
			name: "protobuf length delimited field 1 wire type 2",
			data: []byte{0x0A},
			want: ContentTypeProtobuf,
		},
		{
			name: "protobuf field 2 length delimited",
			data: []byte{0x12},
			want: ContentTypeProtobuf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectContentType(tt.data)
			if got != tt.want {
				t.Errorf("DetectContentType(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   ContentType
	}{
		{
			name:   "empty accept",
			accept: "",
			want:   ContentTypeJSON,
		},
		{
			name:   "application/json",
			accept: "application/json",
			want:   ContentTypeJSON,
		},
		{
			name:   "application/x-protobuf",
			accept: "application/x-protobuf",
			want:   ContentTypeProtobuf,
		},
		{
			name:   "application/vnd.google.protobuf",
			accept: "application/vnd.google.protobuf",
			want:   ContentTypeProtobuf,
		},
		{
			name:   "protobuf with q-value",
			accept: "application/x-protobuf;q=0.9, application/json",
			want:   ContentTypeProtobuf,
		},
		{
			name:   "json only",
			accept: "application/json;q=1",
			want:   ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAcceptHeader(tt.accept)
			if got != tt.want {
				t.Errorf("ParseAcceptHeader(%q) = %v, want %v", tt.accept, got, tt.want)
			}
		})
	}
}

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"application/x-protobuf;q=0.9", "application/x-protobuf", true},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestGetWireType(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantFn  int
		wantWt  int
		wantErr bool
	}{
		{
			name:    "field 1 varint",
			data:    []byte{0x08},
			wantFn:  1,
			wantWt:  0,
			wantErr: false,
		},
		{
			name:    "field 1 length delimited",
			data:    []byte{0x0A},
			wantFn:  1,
			wantWt:  2,
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, wt, err := GetWireType(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWireType(%v) error = %v, wantErr %v", tt.data, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if fn != tt.wantFn || wt != tt.wantWt {
					t.Errorf("GetWireType(%v) = (%d, %d), want (%d, %d)", tt.data, fn, wt, tt.wantFn, tt.wantWt)
				}
			}
		})
	}
}

func TestPeekMessageSize(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantSz  int
		wantOk  bool
	}{
		{
			name:   "length delimited field 1",
			data:   []byte{0x0A, 10},
			wantSz: 10,
			wantOk: true,
		},
		{
			name:   "length delimited field 2",
			data:   []byte{0x12, 5},
			wantSz: 5,
			wantOk: true,
		},
		{
			name:   "too short",
			data:   []byte{0x0A},
			wantOk: false,
		},
		{
			name:   "not length delimited",
			data:   []byte{0x08},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sz, ok := PeekMessageSize(tt.data)
			if ok != tt.wantOk || (ok && sz != tt.wantSz) {
				t.Errorf("PeekMessageSize(%v) = (%d, %v), want (%d, %v)", tt.data, sz, ok, tt.wantSz, tt.wantOk)
			}
		})
	}
}

func TestValidateSchemaVersion(t *testing.T) {
	supported := []uint32{1, 2, 3}

	tests := []struct {
		name    string
		version uint32
		want    bool
	}{
		{"version 1 supported", 1, true},
		{"version 2 supported", 2, true},
		{"version 3 supported", 3, true},
		{"version 0 not supported", 0, false},
		{"version 4 not supported", 4, false},
		{"version 100 not supported", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateSchemaVersion(tt.version, supported)
			if got != tt.want {
				t.Errorf("ValidateSchemaVersion(%d, %v) = %v, want %v", tt.version, supported, got, tt.want)
			}
		})
	}
}