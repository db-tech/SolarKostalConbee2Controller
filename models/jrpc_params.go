package models

const (
	InitStatusOk = iota
	InitStatusDeconzAuth
	InitStatusConfig
	InitStatusError
	InitStatusKostalAuth
	InitStatusLogin
)

type InitResponseParams struct {
	Status        int
	StatusMessage string
}

type RestErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type AuthenticateParams struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	HostAddress string `json:"hostAddress"`
}

type SavePropertiesParams struct {
	Threshold    float64 `json:"Threshold"`
	PlugName     string  `json:"PlugName"`
	PollDuration int     `json:"PollDuration"`
}

type SwitchLightParams struct {
	LightId string `json:"lightId"`
}

type MonitoringEnabledParams struct {
	Enabled bool `json:"enabled"`
}
