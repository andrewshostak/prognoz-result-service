package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/andrewshostak/result-service/errs"
)

const fixturesPath = "/v3/fixtures"
const authHeader = "X-RapidAPI-Key"

type FootballAPIClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewFootballAPIClient(httpClient *http.Client, baseURL string, apiKey string) *FootballAPIClient {
	return &FootballAPIClient{httpClient: httpClient, baseURL: baseURL, apiKey: apiKey}
}

func (c *FootballAPIClient) SearchFixtures(ctx context.Context, search FixtureSearch) (*FixturesResponse, error) {
	url := c.baseURL + fixturesPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get fixtures: %w", err)
	}

	q := req.URL.Query()
	q.Add("season", strconv.Itoa(int(search.Season)))
	q.Add("timezone", search.Timezone)

	if search.TeamID != nil {
		q.Add("team", strconv.Itoa(int(*search.TeamID)))
	}

	if search.Date != nil {
		q.Add("date", *search.Date)
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set(authHeader, c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to get fixtures: %w", err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Printf("couldn't close response body: %v", err)
		}
	}()

	if res.StatusCode == http.StatusOK {
		var body FixturesResponse
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("failed to decode get fixtures response body: %w", err)
		}

		return &body, nil
	}

	return nil, fmt.Errorf("%s: %w", fmt.Sprintf("failed to get fixtures, status %d", res.StatusCode), errs.ErrUnexpectedAPIFootballStatusCode)
}
