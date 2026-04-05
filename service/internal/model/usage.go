package model

// WindowUsage represents a rolling usage window (e.g. five_hour, seven_day).
// UtilizationPct and ResetsAt are pointers to distinguish "not present" from zero.
type WindowUsage struct {
	UtilizationPct *float64 `json:"utilizationPct"`
	ResetsAt       *string  `json:"resetsAt"`
}

type ExtraUsage struct {
	IsEnabled      bool     `json:"isEnabled"`
	MonthlyLimit   int      `json:"monthlyLimit"`
	UsedCredits    float64  `json:"usedCredits"`
	UtilizationPct *float64 `json:"utilizationPct"`
}

type UsagePayload struct {
	Platform          string       `json:"platform"`
	FiveHour          *WindowUsage `json:"fiveHour"`
	SevenDay          *WindowUsage `json:"sevenDay"`
	SevenDayOauthApps *WindowUsage `json:"sevenDayOauthApps"`
	SevenDayOpus      *WindowUsage `json:"sevenDayOpus"`
	SevenDaySonnet    *WindowUsage `json:"sevenDaySonnet"`
	SevenDayCowork    *WindowUsage `json:"sevenDayCowork"`
	IguanaNecktie     *WindowUsage `json:"iguanaNecktie"`
	ExtraUsage        *ExtraUsage  `json:"extraUsage"`
	CapturedAt        string       `json:"capturedAt"`
}
