package storage

import (
	"os"
	"strings"
)

// b2PlaceholderValues are example .env values that must not be sent to B2 (401 / noise).
var b2PlaceholderValues = map[string]struct{}{
	"":                          {},
	"your_account_id":           {},
	"your_application_key":      {},
	"your_bucket_name":          {},
	"your_account_id_here":      {},
	"your_application_key_here": {},
	"your_bucket_name_here":     {},
}

func isB2Placeholder(s string) bool {
	_, ok := b2PlaceholderValues[strings.ToLower(strings.TrimSpace(s))]
	return ok
}

// firstNonEmpty returns the first trimmed non-empty string among candidates.
func firstNonEmpty(candidates ...string) string {
	for _, c := range candidates {
		t := strings.TrimSpace(c)
		if t != "" {
			return t
		}
	}
	return ""
}

// B2ConfigFromEnv reads Backblaze B2 credentials using the same names as the Next.js frontend
// (BACKBLAZE_KEY_ID, BACKBLAZE_APP_KEY, BACKBLAZE_BUCKET_NAME), with fallback to legacy B2_* vars.
//
// The first value is the Application Key ID (often starts with "005"), not the raw account id
// unless you use the master application key — same as b2.NewClient / b2_authorize_account.
func B2ConfigFromEnv() (keyID, applicationKey, bucketName string, ok bool) {
	keyID = firstNonEmpty(os.Getenv("BACKBLAZE_KEY_ID"), os.Getenv("B2_ACCOUNT_ID"))
	applicationKey = firstNonEmpty(os.Getenv("BACKBLAZE_APP_KEY"), os.Getenv("B2_APPLICATION_KEY"))
	bucketName = firstNonEmpty(os.Getenv("BACKBLAZE_BUCKET_NAME"), os.Getenv("B2_BUCKET_NAME"))

	if keyID == "" || applicationKey == "" || bucketName == "" {
		return "", "", "", false
	}
	if isB2Placeholder(keyID) || isB2Placeholder(applicationKey) || isB2Placeholder(bucketName) {
		return "", "", "", false
	}
	return keyID, applicationKey, bucketName, true
}
