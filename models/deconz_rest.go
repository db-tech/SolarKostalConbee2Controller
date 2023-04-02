package models

type Device struct {
	ID                string `json:"id"`
	InternalIPAddress string `json:"internalipaddress"`
	MACAddress        string `json:"macaddress"`
	InternalPort      int    `json:"internalport"`
	Name              string `json:"name"`
	PublicIPAddress   string `json:"publicipaddress"`
}

type SuccessResponse struct {
	Success struct {
		Username string `json:"username"`
	} `json:"success"`
}

type PointSymbol struct{}

type State struct {
	Alert     string     `json:"alert"`
	Bri       int        `json:"bri"`
	ColorMode string     `json:"colormode"`
	Ct        int        `json:"ct"`
	Effect    string     `json:"effect"`
	Hue       int        `json:"hue"`
	On        bool       `json:"on"`
	Reachable bool       `json:"reachable"`
	Sat       int        `json:"sat"`
	Xy        [2]float64 `json:"xy"`
}

type Light struct {
	Etag         string      `json:"etag"`
	HasColor     bool        `json:"hascolor"`
	Manufacturer string      `json:"manufacturer"`
	ModelID      string      `json:"modelid"`
	Name         string      `json:"name"`
	PointSymbol  PointSymbol `json:"pointsymbol"`
	State        State       `json:"state"`
	SWVersion    string      `json:"swversion"`
	Type         string      `json:"type"`
	UniqueID     string      `json:"uniqueid"`
}

type LightsResponse map[string]Light
