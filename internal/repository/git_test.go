package repository

import (
	"net"
	"testing"
)

func TestValidateCloneURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid URLs
		{name: "github https", url: "https://github.com/user/repo", wantErr: false},
		{name: "github http", url: "http://github.com/user/repo", wantErr: false},
		{name: "gitlab https", url: "https://gitlab.com/user/repo", wantErr: false},
		{name: "bitbucket", url: "https://bitbucket.org/user/repo", wantErr: false},

		// SSRF: blocked schemes
		{name: "file scheme", url: "file:///etc/passwd", wantErr: true},
		{name: "ftp scheme", url: "ftp://evil.com/repo", wantErr: true},
		{name: "ssh scheme", url: "ssh://git@github.com/repo", wantErr: true},

		// SSRF: blocked hosts
		{name: "localhost", url: "https://localhost/repo", wantErr: true},
		{name: "loopback ip", url: "https://127.0.0.1/repo", wantErr: true},
		{name: "ipv6 loopback", url: "https://[::1]/repo", wantErr: true},
		{name: "0.0.0.0", url: "https://0.0.0.0/repo", wantErr: true},

		// SSRF: private IPs
		{name: "private 10.x", url: "http://10.0.0.1/repo", wantErr: true},
		{name: "private 172.x", url: "http://172.16.0.1/repo", wantErr: true},
		{name: "private 192.168.x", url: "http://192.168.1.1/repo", wantErr: true},
		{name: "link-local", url: "http://169.254.1.1/repo", wantErr: true},

		// Domain not in allowlist
		{name: "unknown domain", url: "https://evil.com/repo", wantErr: true},
		{name: "internal host", url: "https://internal.company.com/repo", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCloneURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCloneURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.0.1", true},
		{"8.8.8.8", false},
		{"140.82.121.4", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			got := isPrivateIP(ip)
			if got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
