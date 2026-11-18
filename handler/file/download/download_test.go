package downloadHandler

import "testing"

func TestParseRange(t *testing.T) {
	tests := []struct {
		name   string
		header string
		size   uint64
		start  int64
		end    int64
		valid  bool
	}{
		{name: "valid bounded range", header: "bytes=0-99", size: 1000, start: 0, end: 99, valid: true},
		{name: "open ended range", header: "bytes=500-", size: 1000, start: 500, end: 999, valid: true},
		{name: "end beyond size clamped", header: "bytes=0-2000", size: 1000, start: 0, end: 999, valid: true},
		{name: "start beyond size", header: "bytes=1000-", size: 1000, valid: false},
		{name: "start equals size", header: "bytes=1000-1000", size: 1000, valid: false},
		{name: "start greater than end", header: "bytes=100-50", size: 1000, valid: false},
		{name: "negative start", header: "bytes=-1-5", size: 1000, valid: false},
		{name: "negative end", header: "bytes=0--5", size: 1000, valid: false},
		{name: "suffix range unsupported", header: "bytes=-500", size: 1000, valid: false},
		{name: "non numeric", header: "bytes=a-b", size: 1000, valid: false},
		{name: "no dash", header: "bytes=5", size: 1000, valid: false},
		{name: "empty size", header: "bytes=0-", size: 0, valid: false},
		{name: "size exceeds signed range", header: "bytes=0-", size: uint64(^uint64(0)), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, valid := parseRange(tt.header, tt.size)
			if valid != tt.valid {
				t.Fatalf("parseRange(%q, %d) valid = %v, want %v", tt.header, tt.size, valid, tt.valid)
			}
			if !valid {
				return
			}
			if start != tt.start || end != tt.end {
				t.Fatalf("parseRange(%q, %d) = (%d, %d), want (%d, %d)", tt.header, tt.size, start, end, tt.start, tt.end)
			}
		})
	}
}
