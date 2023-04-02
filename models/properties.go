package models

import (
	"fmt"
	"github.com/db-tech/SolarKostalConbee2Controller/ini"
	"github.com/rs/zerolog/log"
	"strconv"
)

type Properties struct {
	HostAddress    string
	ApiKey         string
	Threshold      float64
	PollDuration   int
	PlugName       string
	DeconzUsername string
	DeconzPassword string
	KostalUsername string
	KostalPassword string
	KostalAddress  string
	KostalType     string
}

func (p *Properties) SaveToFile(s string) error {
	log.Info().Msg("Save properties to file")
	err := ini.SavePropertiesToFile(s, p.ToMap())
	if err != nil {
		return err
	}
	return nil
}

func FromMapWithDefaults(m map[string]string) (*Properties, error) {
	properties := &Properties{
		HostAddress:    "",
		ApiKey:         "",
		Threshold:      100000.0,
		PollDuration:   10,
		PlugName:       "",
		DeconzUsername: "",
		DeconzPassword: "",
		KostalUsername: "",
		KostalPassword: "",
		KostalAddress:  "",
		KostalType:     "",
	}

	if threshold, ok := m["Threshold"]; ok {
		float, err := strconv.ParseFloat(threshold, 64)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		properties.Threshold = float
	}

	if hostAddress, ok := m["hostAddress"]; ok {
		properties.HostAddress = hostAddress
	}

	if apiKey, ok := m["apiKey"]; ok {
		properties.ApiKey = apiKey
	}

	if username, ok := m["deconzUsername"]; ok {
		properties.DeconzUsername = username
	}

	if password, ok := m["deconzPassword"]; ok {
		properties.DeconzPassword = password
	}

	if plugName, ok := m["plugName"]; ok {
		properties.PlugName = plugName
	}

	if kostalUsername, ok := m["kostalUsername"]; ok {
		properties.KostalUsername = kostalUsername
	}

	if kostalPassword, ok := m["kostalPassword"]; ok {
		properties.KostalPassword = kostalPassword
	}

	if kostalAddress, ok := m["kostalAddress"]; ok {
		properties.KostalAddress = kostalAddress
	}

	if kostalType, ok := m["kostalType"]; ok {
		properties.KostalType = kostalType
	}

	if pollDuration, ok := m["pollDuration"]; ok {
		pullDurationInt, err := strconv.Atoi(pollDuration)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		properties.PollDuration = pullDurationInt
	}
	return properties, nil
}

func (p *Properties) ToMap() map[string]string {
	return map[string]string{
		"hostAddress":    p.HostAddress,
		"apiKey":         p.ApiKey,
		"deconzUsername": p.DeconzUsername,
		"deconzPassword": p.DeconzPassword,
		"Threshold":      fmt.Sprintf("%f", p.Threshold),
		"plugName":       p.PlugName,
		"pollDuration":   fmt.Sprintf("%d", p.PollDuration),
		"kostalUsername": p.KostalUsername,
		"kostalPassword": p.KostalPassword,
		"kostalAddress":  p.KostalAddress,
		"kostalType":     p.KostalType,
	}
}
