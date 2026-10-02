package command

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AgentDrasil/ssh-wrapper/lib/config"
)

var (
	ErrAccessDenied = errors.New("access denied")

	quotedPathRegex = regexp.MustCompile(`'([^']+)'`)
	hostRegex       = regexp.MustCompile(`@([a-zA-Z0-9\-\.]+)`)
)

func parsePath(cmd string) string {
	matches := quotedPathRegex.FindStringSubmatch(cmd)
	if len(matches) < 2 {
		return ""
	}
	path := matches[1]
	if idx := strings.Index(path, "://"); idx >= 0 {
		path = path[idx+3:]
		if slashIdx := strings.Index(path, "/"); slashIdx >= 0 {
			path = path[slashIdx+1:]
		} else {
			return ""
		}
	} else if idx := strings.Index(path, ":"); idx >= 0 {
		path = path[idx+1:]
	}
	return path
}

func parseHost(cmd string) string {
	matches := hostRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func isHostAllowed(cmd string, conf *config.Config) *config.AllowedRule {
	host := parseHost(cmd)
	if host == "" {
		return nil
	}
	for _, entry := range conf.Allowed {
		if entry.Host == host {
			return &entry
		}
	}
	return nil
}

func IsGitCommand(cmd string) bool {
	return strings.Contains(cmd, "git-")
}

func IsBasicHandshake(cmd string) bool {
	return !IsGitCommand(cmd)
}

func VerifyAccess(cmd string, conf *config.Config) (*config.AllowedRule, error) {
	host := parseHost(cmd)

	rule := isHostAllowed(cmd, conf)
	if rule == nil {
		return nil, fmt.Errorf("%w: host '%s' is not in allowlist", ErrAccessDenied, host)
	}

	// No path_prefix: path matching is skipped for non-git commands.
	// Git commands still need an explicit path_prefix to be allowed.
	if len(rule.PathPrefix) == 0 && !IsGitCommand(cmd) {
		return rule, nil
	}

	path := parsePath(cmd)
	if path != "" {
		for _, prefix := range rule.PathPrefix {
			if strings.HasPrefix(path, prefix) {
				return rule, nil
			}
		}
		return nil, fmt.Errorf("%w: host '%s', path '%s' does not match any allowed path_prefix", ErrAccessDenied, host, path)
	} else if IsBasicHandshake(cmd) {
		return rule, nil
	}

	return nil, ErrAccessDenied
}
