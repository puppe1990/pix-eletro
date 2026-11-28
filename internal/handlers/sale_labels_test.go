package handlers

import "testing"

func TestChannelLabel_Facebook(t *testing.T) {
	if got := channelLabel("facebook"); got != "Facebook" {
		t.Errorf("channelLabel(facebook) = %q, want Facebook", got)
	}
	if !validChannel("facebook") {
		t.Error("validChannel(facebook) = false, want true")
	}
	found := false
	for _, opt := range channelOptions() {
		if opt["value"] == "facebook" && opt["label"] == "Facebook" {
			found = true
		}
	}
	if !found {
		t.Error("channelOptions() is missing the Facebook option")
	}
}
