package practice

import "testing"

func TestNormalizeAudioMIME(t *testing.T) {
	tests := []struct {
		contentType string
		filename    string
		want        string
	}{
		{"audio/webm;codecs=opus", "recording.webm", "audio/webm"},
		{"AUDIO/WEBM; codecs=opus", "x.webm", "audio/webm"},
		{"application/octet-stream", "recording.webm", "audio/webm"},
		{"", "recording.webm", "audio/webm"},
		{"audio/mp4", "voice.m4a", "audio/mp4"},
		{"video/webm", "recording.webm", "video/webm"},
	}

	for _, tt := range tests {
		got := normalizeAudioMIME(tt.contentType, tt.filename)
		if got != tt.want {
			t.Errorf("normalizeAudioMIME(%q, %q) = %q, want %q",
				tt.contentType, tt.filename, got, tt.want)
		}
	}
}

func TestAllowedAudioMimesIncludesWebMVariants(t *testing.T) {
	for _, mime := range []string{"audio/webm", "video/webm"} {
		if !allowedAudioMimes[mime] {
			t.Fatalf("expected %q to be allowed", mime)
		}
	}
}
