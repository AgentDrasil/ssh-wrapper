package command

import (
	"testing"

	"github.com/AgentDrasil/ssh-wrapper/lib/config"
	"github.com/stretchr/testify/assert"
)

// ssh claude@localvm "docker rm helper; docker run ... sh -c 'ls -lRa /data/'"
// main.go gives VerifyAccess os.Args[1:].
func TestVerifyAccessNonGitCommandWithQuotes(t *testing.T) {
	conf := &config.Config{
		Allowed: []config.AllowedRule{
			{Host: "localvm"},
		},
	}
	args := []string{
		"claude@localvm",
		"docker rm helper; docker run -v caddy-cert:/data --name helper busybox sh -c 'ls -lRa /data/'",
	}

	rule, err := VerifyAccess(args, conf)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
}

func TestVerifyAccessNonGitCommandWithDoubleQuotes(t *testing.T) {
	conf := &config.Config{
		Allowed: []config.AllowedRule{
			{Host: "localvm"},
		},
	}
	args := []string{
		"claude@localvm",
		`docker rm helper; docker run -v caddy-cert:/data --name helper busybox sh -c "ls -lRa /data/"`,
	}

	rule, err := VerifyAccess(args, conf)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
}
