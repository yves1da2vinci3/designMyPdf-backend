package fbadmin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	infisical "github.com/infisical/go-sdk"
	"google.golang.org/api/option"
)

// NewAuthClient builds a Firebase Auth client using, in order:
// 1) env FIREBASE_SERVICE_ACCOUNT (raw JSON, local/dev),
// 2) file GOOGLE_APPLICATION_CREDENTIALS,
// 3) Infisical (INFISICAL_ACCESS_TOKEN or universal auth + project + environment).
func NewAuthClient(ctx context.Context) (*fbauth.Client, error) {
	jsonBytes, source, err := loadServiceAccountJSON(ctx)
	if err != nil {
		return nil, err
	}
	if len(jsonBytes) == 0 {
		return nil, fmt.Errorf("empty service account JSON from %s", source)
	}
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("firebase.NewApp: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase Auth: %w", err)
	}
	return client, nil
}

func loadServiceAccountJSON(ctx context.Context) ([]byte, string, error) {
	if v := strings.TrimSpace(os.Getenv("FIREBASE_SERVICE_ACCOUNT")); v != "" {
		return []byte(v), "FIREBASE_SERVICE_ACCOUNT", nil
	}
	if p := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, "", fmt.Errorf("read GOOGLE_APPLICATION_CREDENTIALS: %w", err)
		}
		return b, "GOOGLE_APPLICATION_CREDENTIALS file", nil
	}
	b, err := loadFromInfisical(ctx)
	if err != nil {
		return nil, "", err
	}
	return b, "Infisical secret FIREBASE_SERVICE_ACCOUNT", nil
}

func loadFromInfisical(ctx context.Context) ([]byte, error) {
	siteURL := strings.TrimSpace(os.Getenv("INFISICAL_SITE_URL"))
	if siteURL == "" {
		siteURL = "https://app.infisical.com"
	}
	envName := strings.TrimSpace(os.Getenv("INFISICAL_ENVIRONMENT"))
	projectSlug := strings.TrimSpace(os.Getenv("INFISICAL_PROJECT_SLUG"))
	projectID := strings.TrimSpace(os.Getenv("INFISICAL_PROJECT_ID"))
	if envName == "" || (projectSlug == "" && projectID == "") {
		return nil, errors.New("configure Infisical (INFISICAL_ENVIRONMENT + INFISICAL_PROJECT_SLUG or INFISICAL_PROJECT_ID), or set FIREBASE_SERVICE_ACCOUNT / GOOGLE_APPLICATION_CREDENTIALS")
	}

	accessToken := strings.TrimSpace(os.Getenv("INFISICAL_ACCESS_TOKEN"))
	clientID := strings.TrimSpace(os.Getenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_ID"))
	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv("CLIENT_ID"))
	}
	clientSecret := strings.TrimSpace(os.Getenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET"))
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(os.Getenv("CLIENT_SECRET"))
	}
	if accessToken == "" && (clientID == "" || clientSecret == "") {
		return nil, errors.New("set INFISICAL_ACCESS_TOKEN or INFISICAL_UNIVERSAL_AUTH_CLIENT_ID + INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET")
	}

	ic := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl:          siteURL,
		AutoTokenRefresh: true,
		SilentMode:       true,
	})

	if accessToken != "" {
		ic.Auth().SetAccessToken(accessToken)
	} else {
		if _, err := ic.Auth().UniversalAuthLogin(clientID, clientSecret); err != nil {
			return nil, fmt.Errorf("infisical universal auth: %w", err)
		}
	}

	opts := infisical.RetrieveSecretOptions{
		SecretKey:   "FIREBASE_SERVICE_ACCOUNT",
		Environment: envName,
		SecretPath:  "/",
	}
	if projectID != "" {
		opts.ProjectID = projectID
	} else {
		opts.ProjectSlug = projectSlug
	}

	sec, err := ic.Secrets().Retrieve(opts)
	if err != nil {
		return nil, fmt.Errorf("infisical retrieve FIREBASE_SERVICE_ACCOUNT: %w", err)
	}
	return []byte(sec.SecretValue), nil
}
