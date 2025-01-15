package helper

import (
	"regexp"
	"strings"
)

// GenerateSlug generates a URL-friendly slug from a given name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "")

	return slug
}
