package logic

import "reward/internal/config"

type Client struct {
	*config.Config
}

func New(c *config.Config) *Client {
	return &Client{
		c,
	}
}
