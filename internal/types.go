package internal

type CheckRequest struct {
	Identifier string  `json:"identifier"`
	Capacity   float64 `json:"capacity"`
	RefillRate float64 `json:"refill_rate"`
}

type CheckResponse struct {
	Allowed    bool    `json:"allowed"`
	Remaining  float64 `json:"remaining"`
	Limit      float64 `json:"limit"`
	RetryAfter float64 `json:"retry_after, omitempty"`
}
