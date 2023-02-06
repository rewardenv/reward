package logic

import "github.com/rewardenv/reward/internal/config"

type Client struct {
	*config.Config
}

func New(c *config.Config) *Client {
	return &Client{
		c,
	}
}
