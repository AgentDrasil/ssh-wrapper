package command

import (
	"testing"

	"github.com/AgentDrasil/ssh-wrapper/lib/config"
	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "git clone with colon format",
			args:     []string{"clone", "'git@github.com:user/repo.git'"},
			expected: "user/repo.git",
		},
		{
			name:     "git clone with slash format",
			args:     []string{"clone", "'ssh://git@github.com/user/repo.git'"},
			expected: "user/repo.git",
		},
		{
			name:     "git clone git-server",
			args:     []string{"clone", "'git@git-server:/git-server/repos/repo1'"},
			expected: "/git-server/repos/repo1",
		},
		{
			name:     "path without host",
			args:     []string{"clone", "'/user/repo.git'"},
			expected: "/user/repo.git",
		},
		{
			name:     "path without host no slash",
			args:     []string{"'user/repo.git'"},
			expected: "user/repo.git",
		},
		{
			name:     "no match",
			args:     []string{"some", "other", "command"},
			expected: "",
		},
		{
			name:     "empty string",
			args:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePath(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseHost(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "github.com",
			args:     []string{"ssh", "-T", "git@github.com"},
			expected: "github.com",
		},
		{
			name:     "bitbucket.org",
			args:     []string{"ssh", "-T", "git@bitbucket.org"},
			expected: "bitbucket.org",
		},
		{
			name:     "no host",
			args:     []string{"git-upload-pack", "user/repo"},
			expected: "",
		},
		{
			name:     "empty string",
			args:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHost(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGitCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "contains git-upload-pack",
			args:     []string{"git-upload-pack", "user/repo"},
			expected: true,
		},
		{
			name:     "contains git-receive-pack",
			args:     []string{"git-receive-pack", "user/repo"},
			expected: true,
		},
		{
			name:     "regular ssh command",
			args:     []string{"ssh", "git@github.com"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitCommand(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBasicHandshake(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "non-git command",
			args:     []string{"ssh", "-T", "git@github.com"},
			expected: true,
		},
		{
			name:     "git command",
			args:     []string{"git-upload-pack", "user/repo"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBasicHandshake(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVerifyAccess(t *testing.T) {
	conf := &config.Config{
		Allowed: []config.AllowedRule{
			{
				Host:       "github.com",
				PathPrefix: []string{"user1/", "user2/"},
			},
			{
				Host:       "ghhy",
				Hostname:   "github.com",
				KeyPath:    "/etc/keys/ghhy_key",
				PathPrefix: []string{"user1/"},
			},
		},
	}

	tests := []struct {
		name         string
		args         []string
		expectErr    bool
		errInclusive string
		expectedKey  string
	}{
		{
			name:      "allowed namespace",
			args:      []string{"clone", "'git@github.com:user1/repo.git'"},
			expectErr: false,
		},
		{
			name:        "allowed namespace with alias ghhy",
			args:        []string{"clone", "'git@ghhy:user1/repo.git'"},
			expectErr:   false,
			expectedKey: "/etc/keys/ghhy_key",
		},
		{
			name:         "disallowed path prefix",
			args:         []string{"clone", "'git@github.com:user3/repo.git'"},
			expectErr:    true,
			errInclusive: "path_prefix",
		},
		{
			name:         "disallowed host namespace",
			args:         []string{"clone", "'git@ggithub.com:user1/repo.git'"},
			expectErr:    true,
			errInclusive: "ggithub.com",
		},
		{
			name:      "basic handshake allowed",
			args:      []string{"ssh", "-T", "git@github.com"},
			expectErr: false,
		},
		{
			name:         "disallowed host handshake",
			args:         []string{"ssh", "-T", "git@ggithub.com"},
			expectErr:    true,
			errInclusive: "ggithub.com",
		},
		{
			name:         "git command without namespace",
			args:         []string{"git-upload-pack", "user3/repo"},
			expectErr:    true,
			errInclusive: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := VerifyAccess(tt.args, conf)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errInclusive)
			} else {
				assert.NoError(t, err)
				if tt.expectedKey != "" {
					assert.Equal(t, tt.expectedKey, rule.KeyPath)
				}
			}
		})
	}
}
