package command

import (
	"testing"

	"github.com/AgentDrasil/ssh-wrapper/lib/config"
	"github.com/stretchr/testify/assert"
)

// On a host without path_prefix, non-git commands are allowed as soon as the
// host matches (path matching is skipped). Git commands stay denied: a repo
// must be explicitly allowed with a path_prefix.
func TestVerifyAccessHostWithoutPathPrefix(t *testing.T) {
	conf := &config.Config{
		Allowed: []config.AllowedRule{
			{Host: "localvm"},
			{Host: "github.com", PathPrefix: []string{"user1/"}},
		},
	}

	tests := []struct {
		name         string
		args         []string
		expectErr    bool
		errInclusive string
	}{
		{
			name: "non-git command with single quotes",
			args: []string{
				"claude@localvm",
				"docker rm helper; docker run -v caddy-cert:/data --name helper busybox sh -c 'ls -lRa /data/'",
			},
		},
		{
			name: "non-git command with double quotes",
			args: []string{
				"claude@localvm",
				`docker rm helper; docker run -v caddy-cert:/data --name helper busybox sh -c "ls -lRa /data/"`,
			},
		},
		{
			name: "non-git command without quotes",
			args: []string{"claude@localvm", "ls", "-l", "/data"},
		},
		{
			name:         "git command with quoted path is denied",
			args:         []string{"git@localvm", "git-upload-pack 'anyone/repo.git'"},
			expectErr:    true,
			errInclusive: "path_prefix",
		},
		{
			name:         "git command without parseable path is denied",
			args:         []string{"git@localvm", "git-upload-pack", "anyone/repo.git"},
			expectErr:    true,
			errInclusive: "access denied",
		},
		{
			name:         "unknown host is still denied",
			args:         []string{"claude@othervm", "docker run busybox sh -c 'ls /data/'"},
			expectErr:    true,
			errInclusive: "othervm",
		},
		{
			name:         "host with path_prefix is still restricted",
			args:         []string{"git@github.com", "git-upload-pack 'user3/repo.git'"},
			expectErr:    true,
			errInclusive: "path_prefix",
		},
		{
			name: "host with path_prefix and allowed path",
			args: []string{"git@github.com", "git-upload-pack 'user1/repo.git'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := VerifyAccess(tt.args, conf)
			if tt.expectErr {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errInclusive)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, rule)
		})
	}
}
