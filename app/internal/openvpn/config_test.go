package openvpn

import "testing"

func TestClientConfigPointsAtServer1Only(t *testing.T) {
	cfg := ClientConfig("185.10.20.30", 1194, "phone")
	if ContainsForbiddenRemote(cfg, "89.20.30.40") {
		t.Fatal("config must not contain server2 public IP")
	}
	if !contains(cfg, "remote 185.10.20.30 1194") {
		t.Fatal("expected remote server1")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && (stringIndex(s, sub) >= 0)))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
