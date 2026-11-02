package ai

import "testing"

func TestAudioUploadMeta(t *testing.T) {
	tests := []struct {
		filename string
		mime     string
		wantName string
		wantMIME string
	}{
		{"recording.webm", "audio/webm;codecs=opus", "recording.webm", "audio/webm"},
		{"recording.webm", "audio/webm", "recording.webm", "audio/webm"},
		{"recording", "audio/webm", "recording.webm", "audio/webm"},
		{"", "video/webm", "audio.webm", "audio/webm"},
		{"voice", "audio/mp4", "voice.m4a", "audio/mp4"},
	}

	for _, tt := range tests {
		name, mime := audioUploadMeta(tt.filename, tt.mime)
		if name != tt.wantName || mime != tt.wantMIME {
			t.Errorf("audioUploadMeta(%q, %q) = (%q, %q), want (%q, %q)",
				tt.filename, tt.mime, name, mime, tt.wantName, tt.wantMIME)
		}
	}
}
