package ratelimit

// Limits defines rate limits per minute and per month.
type Limits struct {
	PerMinute int // requests per minute
	PerMonth  int // requests per month (0 means unlimited)
}

const (
	// Monthly limit for view_collab seats
	viewCollabMonthlyLimit = 6
)

// Per-minute limits from Figma API tiers
var perMinuteLimits = map[int]int{
	1: 120,
	2: 300,
	3: 600,
	4: 1200,
}

// GetLimits returns the rate limits for the given tier and seat type.
// Tier must be 1-4, seatType must be "view_collab" or "dev_full".
// Defaults to tier 1 and dev_full for invalid inputs.
func GetLimits(tier int, seatType string) Limits {
	// Validate tier, default to 1 if invalid
	if tier < 1 || tier > 4 {
		tier = 1
	}
	perMinute := perMinuteLimits[tier]

	// Monthly limits: view_collab seats have 6/month, dev_full unlimited (0)
	perMonth := 0
	if seatType == "view_collab" {
		perMonth = viewCollabMonthlyLimit
	}
	// If seatType is not recognized, treat as dev_full (unlimited)
	return Limits{
		PerMinute: perMinute,
		PerMonth:  perMonth,
	}
}
