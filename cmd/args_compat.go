package cmd

import "strings"

// normalizeProfileAliasArgs rewrites "-profile" to "--profile".
// This accepts single-dash long-option style while keeping Cobra flag definitions standard.
func normalizeProfileAliasArgs(args []string) []string {
	out := make([]string, 0, len(args))

	for _, token := range args {
		if token == "-profile" {
			out = append(out, "--profile")
			continue
		}
		if strings.HasPrefix(token, "-profile=") {
			out = append(out, "--profile="+strings.TrimPrefix(token, "-profile="))
			continue
		}
		out = append(out, token)
	}

	return out
}
