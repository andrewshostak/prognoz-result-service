package handler

type CreateMatchRequest struct {
	StartsAt  string `binding:"required" json:"starts_at"`
	AliasHome string `binding:"required" json:"alias_home"`
	AliasAway string `binding:"required" json:"alias_away"`
}

type CreateSubscriptionRequest struct {
	MatchID   string `binding:"required" json:"match_id"`
	URL       string `binding:"required" json:"url"`
	SecretKey string `binding:"required" json:"secret_key"`
}
