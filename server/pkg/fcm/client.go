package fcm

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	msg *messaging.Client
}

type Config struct {
	CredentialsFile string
	ProjectID       string
}

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

func (c *Client) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	_, err := c.msg.Send(ctx, &messaging.Message{
		Token:        token,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
	})
	return err
}

func IsTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	return messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err)
}

func (c *Client) IsTokenInvalid(err error) bool { return IsTokenInvalid(err) }
