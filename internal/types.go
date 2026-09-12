package internal

type Stats struct {
	Total int64 `json:"total"`
	Allowed int64 `json:"allowed"`
	Denied int64 `json:"denied"`
	RejectionRate string `json:"rejection_rate"`
    ActiveClients int `json:"active_clients"`
    PerClient map[string]ClientStats `json:"per_client"`

}
type ClientStats struct {
    Allowed int64 `json:"allowed"`
    Denied int64 `json:"denied"`
    CurrentTokens float64 `json:"current_tokens"`
}