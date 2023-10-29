package client

import "context"

type FootballAPIClient struct {
	APIKey   string
	Timezone string
}

func NewFootballAPIClient(apiKey string, timezone string) *FootballAPIClient {
	return &FootballAPIClient{APIKey: apiKey, Timezone: timezone}
}

func (c *FootballAPIClient) SearchFixtures(ctx context.Context, search FixtureSearch) (*FixturesResponse, error) {
	return nil, nil
}
