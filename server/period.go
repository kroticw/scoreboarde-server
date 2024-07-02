package server

type Period struct {
	TimeStart    int64 `json:"time_start"`
	Count        int   `json:"count"`
	TimeInPeriod int64 `json:"time_in_period"`
	Pause        bool  `json:"pause"`
}
