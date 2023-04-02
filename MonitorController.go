package main

import (
	"errors"
	"github.com/db-tech/JsonRpcWebsocketServer/jrws"
	"github.com/db-tech/SolarKostalConbee2Controller/models"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	MonitoringEventPropertiesUpdated = "PropertiesUpdated"
	MonitoringEventStopMonitoring    = "StopMonitoring"
)

type MonitoringController struct {
	conbeeClient    *ConbeeClient
	inverter        Inverter
	websocketServer *jrws.WebsocketServer
	eventChan       chan string
	startEventChan  chan models.Properties
	isRunning       bool
}

func NewMonitoringController(conbeeClient *ConbeeClient, inverter Inverter, websocketServer *jrws.WebsocketServer) *MonitoringController {
	monitoring := &MonitoringController{
		conbeeClient:    conbeeClient,
		inverter:        inverter,
		websocketServer: websocketServer,
		eventChan:       make(chan string),
		startEventChan:  make(chan models.Properties),
		isRunning:       false,
	}
	monitoring.run()
	return monitoring
}

type Data struct {
	InverterData InverterData `json:"inverterData"`
	SocketState  bool         `json:"socketState"`
}

func (m *MonitoringController) RequestDataAndSendWsNotification(properties models.Properties) (Data, error) {
	log.Info().Msg("Get inverter data")
	if m.inverter == nil {
		log.Error().Msg("Inverter is nil")
		return Data{}, errors.New("inverter is nil")
	}
	inverterData, err := m.inverter.GetInverterData()
	if err != nil {
		log.Error().Err(err).Msg("Could not get inverter data")
		return Data{}, err
	}

	socketState, err := m.conbeeClient.IsLightOn(properties.PlugName)
	if err != nil {
		log.Error().Err(err).Msg("Could not get socket state")
		return Data{}, err
	}

	data := Data{
		InverterData: inverterData,
		SocketState:  socketState,
	}

	err = m.websocketServer.WriteNotificationToAllMembers("data", data)
	if err != nil {
		log.Error().Err(err).Msg("Could not write notification to all members")
		return data, err
	}
	return data, nil
}

func (m *MonitoringController) run() {
	go func() {
		defer func() {
			log.Info().Msg("MonitoringController: stopped monitoring")
			m.isRunning = false
			m.websocketServer.WriteNotificationToAllMembers("monitoring", models.MonitoringEnabledParams{Enabled: false})
		}()
		log.Info().Msg("MonitoringController: started monitoring")
		properties := models.Properties{
			PollDuration: 10,
		}
		log.Info().Msgf("MonitoringController set poll duration to %d", properties.PollDuration)
		ticker := time.NewTicker(time.Duration(properties.PollDuration) * time.Second)
		ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Info().Msg("MonitoringController: tick")
				data, err := m.RequestDataAndSendWsNotification(properties)
				if err != nil {
					log.Error().Err(err).Msg("Could not request data and send ws notification")
					return
				}
				housePowerConsumption := data.InverterData.HousePowerConsumption
				if data.SocketState == true {
					housePowerConsumption -= properties.Threshold
				}

				if data.InverterData.Overproduction > properties.Threshold {
					log.Info().Msgf("Grid consumption: %f, so switching lights on", data.InverterData.Overproduction)
					err := m.conbeeClient.SwitchOnLight(properties.PlugName)
					if err != nil {
						log.Error().Err(err).Msg("Could not switch on light")
						return
					}
				} else if data.InverterData.Overproduction < 0 {
					log.Info().Msgf("Grid consumption: %f, so switching lights off", data.InverterData.Overproduction)
					err := m.conbeeClient.SwitchOffLight(properties.PlugName)
					if err != nil {
						log.Error().Err(err).Msg("Could not switch off light")
						return
					}
				}
				data.SocketState, err = m.conbeeClient.IsLightOn(properties.PlugName)
				if err != nil {
					log.Error().Err(err).Msg("Could not get socket state")
					return
				}
				err = m.websocketServer.WriteNotificationToAllMembers("data", data)
				if err != nil {
					log.Error().Err(err).Msg("Could not write notification to all members")
					return
				}
			case event := <-m.eventChan:
				switch event {
				case MonitoringEventStopMonitoring:
					ticker.Stop()
					log.Info().Msg("Stop monitoring")
					m.isRunning = false
					m.websocketServer.WriteNotificationToAllMembers("monitoring", models.MonitoringEnabledParams{Enabled: false})
				default:
					log.Error().Msgf("Unknown event: %s", event)
				}
			case properties = <-m.startEventChan:
				log.Info().Msg("Start monitoring")
				m.isRunning = true
				ticker.Reset(time.Duration(properties.PollDuration) * time.Second)
				m.websocketServer.WriteNotificationToAllMembers("monitoring", models.MonitoringEnabledParams{Enabled: true})
			}

		}
	}()
}

func (m *MonitoringController) StopMonitoring() {
	m.eventChan <- MonitoringEventStopMonitoring
}

func (m *MonitoringController) StartMonitoring(properties models.Properties) {
	m.startEventChan <- properties
}

func (m *MonitoringController) IsRunning() bool {
	return m.isRunning
}
