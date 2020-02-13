package component

import (
	"testing"
)

func TestsshPubKeys(t *testing.T) {
	setting := SSHConfig{
		TempDir:              "/tmp/.ssh",
		UserDir:              "/tmp/user",
		DefaultSShPubKeyFile: "ssh/pub_key",
	}
	sshPubKeys(setting)
}
