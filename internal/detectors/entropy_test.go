package detectors

import "testing"

func TestEntropyNearbySecrets_Negative_BearerHeader(t *testing.T) {
	// Authorization bearer headers without other context should not trigger now
	data := []byte("Authorization: Bearer abcdefghijklmnopqrstuvwxyz0123456789AB")
	fs := EntropyNearbySecrets("hdr.txt", data)
	if len(fs) != 0 {
		t.Fatalf("unexpected finding for bearer header context")
	}
}
