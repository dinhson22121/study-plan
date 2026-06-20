// Package fcm wraps the Firebase Admin SDK messaging client. It exposes a small
// Send surface plus invalid-token detection; higher-level retry/backoff and
// domain mapping live in the notification module's FCM adapter.
package fcm

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client is a thin wrapper over the Firebase messaging client.
type Client struct {
	msg *messaging.Client
}

// Config holds Firebase initialization parameters.
type Config struct {
	CredentialsFile string
	ProjectID       string
}

// New initializes the Firebase app from a service-account credentials file and
// returns a messaging Client. ProjectID is optional when present in the creds.
func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.CredentialsFile == "" {
		return nil, errors.New("fcm: credentials file path is required")
	}
	opts := []option.ClientOption{option.WithCredentialsFile(cfg.CredentialsFile)}
	fbCfg := &firebase.Config{}
	if cfg.ProjectID != "" {
		fbCfg.ProjectID = cfg.ProjectID
	}
	app, err := firebase.NewApp(ctx, fbCfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("fcm: init firebase app: %w", err)
	}
	msg, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcm: init messaging client: %w", err)
	}
	return &Client{msg: msg}, nil
}

// Send delivers a single notification to a device token with optional data
// payload. It returns the raw Firebase error on failure so callers can classify
// it via IsTokenInvalid.
func (c *Client) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	_, err := c.msg.Send(ctx, &messaging.Message{
		Token:        token,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
	})
	return err
}

// IsTokenInvalid reports whether a Firebase send error indicates the token is
// unregistered or invalid, so the caller can deactivate it instead of retrying.
func IsTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	return messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err)
}

// IsTokenInvalid is the method form, letting *Client satisfy interfaces that
// require both Send and invalid-token classification.
func (c *Client) IsTokenInvalid(err error) bool { return IsTokenInvalid(err) }
