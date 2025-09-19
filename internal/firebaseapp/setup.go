package firebaseapp

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// New initializes a Firebase App and Auth client.
// If credentialsFile is provided, it uses the specified credentials file.
// If credentialsFile is empty, it uses Application Default Credentials (ADC).
func New(ctx context.Context, credentialsFile string) (*firebase.App, *firebaseauth.Client, error) {
	var app *firebase.App
	var err error

	if credentialsFile == "" {
		// Use Application Default Credentials (ADC)
		app, err = firebase.NewApp(ctx, nil)
	} else {
		// Use specified credentials file
		opt := option.WithCredentialsFile(credentialsFile)
		app, err = firebase.NewApp(ctx, nil, opt)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("firebase: failed to initialize app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("firebase: failed to init auth client: %w", err)
	}

	return app, authClient, nil
}
