package main

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// InitFirebase initializes the Firebase app and returns an Auth client.
// Supports either FIREBASE_SERVICE_ACCOUNT_PATH (file path) or
// FIREBASE_SERVICE_ACCOUNT (raw JSON) environment variables.
func InitFirebase(ctx context.Context) (*firebase.App, *auth.Client, error) {
	svcPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	svcJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT")

	var app *firebase.App
	var err error

	if svcPath != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsFile(svcPath))
		if err != nil {
			return nil, nil, fmt.Errorf("firebase.NewApp (file): %w", err)
		}
	} else if svcJSON != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(svcJSON)))
		if err != nil {
			return nil, nil, fmt.Errorf("firebase.NewApp (json): %w", err)
		}
	} else {
		app, err = firebase.NewApp(ctx, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("firebase.NewApp (default): %w", err)
		}
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("app.Auth: %w", err)
	}

	return app, authClient, nil
}
