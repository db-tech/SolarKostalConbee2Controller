package main

import (
	"encoding/json"
	"fmt"
	"github.com/db-tech/SolarKostalConbee2Controller/models"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

const (
	ApiPath = "/api"
)

type ConbeeClient struct {
	username    string
	password    string
	hostAddress string
	apiKey      string
	restClient  *resty.Client
}

/**
 * Create a new ConbeeClient
 * No API key is set, so you have to either set it manually or call CreateApiKey()
 */
func NewConbeeClient(username string, password string, hostAddress string, apiKey string) *ConbeeClient {
	conbeeClient := &ConbeeClient{
		username:    username,
		password:    password,
		hostAddress: hostAddress,
		apiKey:      apiKey,
		restClient:  resty.New(),
	}
	conbeeClient.restClient.SetBaseURL("http://" + hostAddress)
	return conbeeClient
}

func (c *ConbeeClient) CreateApiKey() (string, error) {
	log.Info().Msg("Creating API key")
	if c.username != "" && c.password != "" {
		log.Debug().Msgf("Trying to authenticate with username %s and address %s ", c.username, c.hostAddress)
		c.restClient.SetBasicAuth(c.username, c.password)
	} else {
		log.Info().Msg("Password or Username not set, create api key manually")
	}

	log.Info().Msg("Send POST request to create API key")
	response, err := c.restClient.R().
		SetBody(fmt.Sprintf(`{"devicetype":"%s"}`, AppName)).
		Post(ApiPath)
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Msg("Error creating API key")
		return "", err
	}
	if response.StatusCode() != 200 {
		log.Error().Stack().Msgf("unexpected return value %d: %v", response.StatusCode(), response.String())
		return "", fmt.Errorf("unexpected return value %d: %v", response.StatusCode(), response.String())
	}

	log.Info().Msg("Successfully created API key")
	// Unmarshal JSON response string into Go struct
	var successResponse []models.SuccessResponse
	err = json.Unmarshal(response.Body(), &successResponse)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}
	if len(successResponse) == 0 {
		log.Fatal().Stack().Msg("Error: " + response.String())
	}

	log.Debug().Msgf("API key: %s", successResponse[0].Success.Username)
	c.apiKey = successResponse[0].Success.Username
	return successResponse[0].Success.Username, nil
}

func (c *ConbeeClient) CheckApiKey() (bool, error) {
	if "" == strings.TrimSpace(c.apiKey) {
		log.Info().Msg("API key not set while checking if API key is valid")
		return false, errors.New("API key not set")
	}
	log.Info().Msgf("Checking API key %s", c.apiKey)
	response, err := c.restClient.R().Get("/api/" + c.apiKey + "/lights")
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Msg("Error checking API url: " + "/api/" + c.apiKey + "/lights")
		return false, err
	}
	if response.StatusCode() == 200 {
		log.Info().Msg("Received 200 OK, API key is valid")
		return true, nil
	}
	if response.StatusCode() == 403 || response.StatusCode() == 401 {
		log.Info().Msgf("Received Status Code %d, API key is invalid", response.StatusCode())
		return false, nil
	}
	return false, fmt.Errorf("unexpected status code: %d", response.StatusCode())
}

func (c *ConbeeClient) GetLights() (map[string]models.Light, *models.RestErrorResponse, error) {
	log.Info().Msgf("Getting lights with API key %s", c.apiKey)
	response, err := c.restClient.R().Get("/api/" + c.apiKey + "/lights")
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode() != 200 {
		return nil, &models.RestErrorResponse{
			Code:    response.StatusCode(),
			Message: response.String(),
		}, nil
	}

	var lights map[string]models.Light
	err = json.Unmarshal(response.Body(), &lights)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, nil, err
	}
	return lights, nil, nil
}

func (c *ConbeeClient) SwitchOnLight(s string) error {
	log.Info().Msgf("Switching on light %s", s)
	id, err := c.GetLightIdByName(s)
	response, err := c.restClient.R().
		SetBody(`{"on":true}`).
		Put("/api/" + c.apiKey + "/lights/" + id + "/state")
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return nil
}

func (c *ConbeeClient) SwitchOffLight(s string) error {
	log.Info().Msgf("Switching off light %s", s)
	id, err := c.GetLightIdByName(s)
	response, err := c.restClient.R().
		SetBody(`{"on":false}`).
		Put("/api/" + c.apiKey + "/lights/" + id + "/state")
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return nil
}

func (c *ConbeeClient) GetLightIdByName(lightName string) (string, error) {
	lights, m, err := c.GetLights()
	if err != nil {
		return "", err
	}
	if m != nil {
		return "", fmt.Errorf("unexpected status code: %d", m.Code)
	}
	for id, light := range lights {
		if light.Name == lightName {
			return id, nil
		}
	}
	return "", fmt.Errorf("light %s not found", lightName)
}

func (c *ConbeeClient) IsLightOn(plugName string) (bool, error) {
	lights, m, err := c.GetLights()
	if err != nil {
		return false, err
	}
	if m != nil {
		return false, fmt.Errorf("unexpected status code: %d", m.Code)
	}
	for _, light := range lights {
		if light.Name == plugName {
			return light.State.On, nil
		}
	}
	return false, fmt.Errorf("light %s not found", plugName)
}

func DiscoverDeconzHostAddress() (string, error) {
	log.Info().Msg("Discovering deconz host address")
	client := resty.New()
	log.Info().Msg("Setting base url to https://phoscon.de")
	log.Warn().Msg("This is a hack to get the deconz host address. It will only work if you have a deconz gateway on your network and an internet connection.")
	client.SetBaseURL("https://phoscon.de")
	response, err := client.R().Get("/discover")
	if err != nil {
		return "", err
	}

	var devices []models.Device

	err = json.Unmarshal(response.Body(), &devices)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	if len(devices) == 0 {
		return "", errors.New("no devices found")
	}

	return devices[0].InternalIPAddress + ":" + strconv.Itoa(devices[0].InternalPort), nil
}
