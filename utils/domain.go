package utils

import (
	"fmt"

	"net/url"

	"regexp"
	"strings"
)

// validateDomainFormat validates if a string is a valid domain name
func validateDomainFormat(domain string) bool {
	// Basic domain regex - matches valid domain names
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(domain) && len(domain) <= 253
}

// StripProtocolFromDomain removes protocol (http://, https://) from domain string
// Returns the clean domain name and an error if the resulting domain is invalid
func StripProtocolFromDomain(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("domain cannot be empty")
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// If it looks like a URL, parse it
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		parsedURL, err := url.Parse(input)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}
		input = parsedURL.Host
	}

	// Remove port if present
	if colonIndex := strings.LastIndex(input, ":"); colonIndex != -1 {
		// Make sure it's actually a port and not part of IPv6
		if !strings.Contains(input, "[") || strings.Contains(input[colonIndex:], "]") {
			input = input[:colonIndex]
		}
	}

	// Validate the resulting domain
	if !validateDomainFormat(input) {
		return "", fmt.Errorf("invalid domain format: %s", input)
	}

	return input, nil
}

// ValidateDomains validates a slice of domain strings, stripping protocols
// Returns clean domains and any validation errors
func ValidateDomains(domains []string) ([]string, error) {
	cleanDomains := make([]string, 0, len(domains))
	var errors []string

	for _, domain := range domains {
		if domain == "" {
			continue // Skip empty domains
		}

		cleanDomain, err := StripProtocolFromDomain(domain)
		if err != nil {
			errors = append(errors, fmt.Sprintf("domain '%s': %v", domain, err))
			continue
		}

		cleanDomains = append(cleanDomains, cleanDomain)
	}

	if len(errors) > 0 {
		return cleanDomains, fmt.Errorf("domain validation errors: %s", strings.Join(errors, "; "))
	}

	return cleanDomains, nil
}
func NormalizeDomain(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("domain cannot be empty")
	}
	input = strings.TrimSpace(input)

	// 如果是 URL，StripProtocolFromDomain 会解析并返回 Host（并校验）
	clean, err := StripProtocolFromDomain(input)
	if err != nil {
		return "", err
	}

	// 去尾随点并转小写
	clean = strings.TrimSuffix(clean, ".")
	clean = strings.ToLower(clean)

	// 最终再校验一次以防异常情况
	if !validateDomainFormat(clean) {
		return "", fmt.Errorf("invalid domain format after normalization: %s", clean)
	}

	return clean, nil
}
func sanitizeFilename(p string) string {
	if p == "/" || p == "" {
		return "root"
	}
	s := strings.Trim(p, "/")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return s
}

// Temporarily commented out for migration
/*
func backupConfig(fc *fastcaddy.FastCaddy, path string) error {
	// 创建备份目录
	if err := os.MkdirAll("backups", 0755); err != nil {
		return err
	}

	cfg, err := fc.GetConfig(path)
	if err != nil {
		// 如果某些路径不存在，仍写入空文件以做记录
		empty := map[string]interface{}{"error": err.Error()}
		b, _ := json.MarshalIndent(empty, "", "  ")
		fname := fmt.Sprintf("%s_%s.json", sanitizeFilename(path), time.Now().Format("20060102T150405"))
		return os.WriteFile(filepath.Join("backups", fname), b, 0644)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	fname := fmt.Sprintf("%s_%s.json", sanitizeFilename(path), time.Now().Format("20060102T150405"))
	return os.WriteFile(filepath.Join("backups", fname), b, 0644)
}

func backupBefore(fc *fastcaddy.FastCaddy, paths ...string) {
	for _, p := range paths {
		if err := backupConfig(fc, p); err != nil {
			log.Printf("备份 %s 失败: %v", p, err)
		} else {
			log.Printf("已备份 %s", p)
		}
	}
}
*/
