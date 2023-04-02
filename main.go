package main

import (
	"encoding/base64"
	"fmt"
	"github.com/db-tech/JsonRpcWebsocketServer/jrws"
	models2 "github.com/db-tech/JsonRpcWebsocketServer/models"
	"github.com/db-tech/SolarKostalConbee2Controller/ini"
	"github.com/db-tech/SolarKostalConbee2Controller/models"
	"github.com/db-tech/SolarKostalConbee2Controller/web"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"time"
)

const (
	AppName    = "Conbee2 Controller"
	AppVersion = "0.0.1"
)

func CreateIniFileIfNotExists(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		_, err := os.Create(filepath)
		if err != nil {
			return err
		}
	}
	return nil
}

func createUsernamePasswordBase64Hash(username string, password string) string {
	usernamePassword := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(usernamePassword))
}

func loadProperties() (*models.Properties, error) {
	log.Info().Msg("Initializing properties")
	err := CreateIniFileIfNotExists("config.ini")
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Load properties from file")
	props, err := ini.LoadPropertiesFromFile("config.ini")
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Create properties from map")
	propertiesWithDefaults, err := models.FromMapWithDefaults(props)
	if err != nil {
		return nil, err
	}
	return propertiesWithDefaults, nil
}

func initDeconzHostAddress() (string, error) {
	hostAddress, err := DiscoverDeconzHostAddress()
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error finding host address")
		return "", err
	}
	return hostAddress, nil
}

func authenticateConbeeClient(properties *models.Properties, conbeeClient *ConbeeClient) models.InitResponseParams {

	if properties.ApiKey == "" {
		log.Warn().Msg("Api key is empty")

		if properties.DeconzUsername != "" && properties.DeconzPassword != "" {
			log.Debug().Msg("Username and password are set, try to create api key")
			apiKEy, err := conbeeClient.CreateApiKey()
			if err != nil {
				return models.InitResponseParams{
					Status: models.InitStatusDeconzAuth,
					StatusMessage: "Although you have set a username and password, " +
						"the api key could not be created. " +
						"Please go to your deconz web interface, " +
						"open the settings page and press the Authenticate button. " +
						"Or provide a valid username and password"}
			}
			properties.ApiKey = apiKEy
			saveErr := properties.SaveToFile("config.ini")
			if saveErr != nil {
				log.Error().Stack().Err(errors.WithStack(saveErr)).Msg("Error saving API key to file")
				return models.InitResponseParams{
					Status:        models.InitStatusError,
					StatusMessage: saveErr.Error(),
				}
			}
		} else {
			apiKey, err := conbeeClient.CreateApiKey()
			if err != nil {
				return models.InitResponseParams{
					Status: models.InitStatusDeconzAuth,
					StatusMessage: "You need to authenticate first. " +
						"Please go to your deconz web interface, " +
						"open the settings page and press the Authenticate button. ",
				}
			}
			properties.ApiKey = apiKey
			saveErr := properties.SaveToFile("config.ini")
			if saveErr != nil {
				log.Error().Stack().Err(errors.WithStack(saveErr)).Msg("Error saving API key to file")
				return models.InitResponseParams{
					Status:        models.InitStatusError,
					StatusMessage: saveErr.Error(),
				}
			}
		}
	}
	_, err := conbeeClient.CheckApiKey()
	if err != nil {
		return models.InitResponseParams{
			Status:        models.InitStatusDeconzAuth,
			StatusMessage: err.Error(),
		}
	}

	_, errResp, err := conbeeClient.GetLights()
	if err != nil {
		return models.InitResponseParams{
			Status:        models.InitStatusDeconzAuth,
			StatusMessage: err.Error(),
		}
	}
	if errResp != nil {
		log.Debug().Msgf("Error response: %v", errResp)
		if errResp.Code == 401 || errResp.Code == 403 {
			log.Debug().Msg("Api key is invalid")
			return models.InitResponseParams{
				Status: models.InitStatusDeconzAuth,
				StatusMessage: "You need to authenticate first. " +
					"Please go to your deconz web interface, " +
					"open the settings page and press the Authenticate button. ",
			}
		}
	}

	return models.InitResponseParams{
		Status:        models.InitStatusOk,
		StatusMessage: "Everything is fine",
	}
}

//go:generate yarn --cwd web dev
//go:generate go build

func printFilesOfCurrentDirectory() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error reading current directory")
		return
	}

	for _, f := range files {
		log.Info().Msg(f.Name())
	}
}

var (
	conbeeClient *ConbeeClient
	inverter     Inverter
	properties   *models.Properties
	monitoring   *MonitoringController
)

func InitConbeeInverterAndProperties() error {
	log.Info().Msg("Starting application")
	var startupStatus models.InitResponseParams
	var err error
	log.Info().Msg("Initializing properties..")
	properties, err = loadProperties()
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error initializing properties")
		return err
	}
	if properties.HostAddress == "" {
		log.Info().Msg("Host address is empty, try to find it")
		hostAddress, err := DiscoverDeconzHostAddress()
		if err != nil {
			log.Error().Stack().Err(err).Msg("Error finding host address")
			return err
		}
		properties.HostAddress = hostAddress
		err = properties.SaveToFile("config.ini")
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("Error saving host address to file")
			return err
		}
	}

	log.Info().Msg("Initializing conbee client..")
	conbeeClient = NewConbeeClient(properties.DeconzUsername, properties.DeconzPassword, properties.HostAddress, properties.ApiKey)

	if !(properties.ApiKey == "" && properties.DeconzUsername == "" && properties.DeconzPassword == "") {
		startupStatus = authenticateConbeeClient(properties, conbeeClient)
		if startupStatus.Status != models.InitStatusOk {
			log.Error().Stack().Err(err).Msg("Error authenticating conbee client")
			return nil
		}
	}

	log.Info().Msg("Initializing inverter..")
	inverter = NewInverterClient(properties.KostalAddress, properties.KostalPassword, properties.KostalType)
	err = inverter.Connect()
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error connecting to inverter")
		return err
	}

	log.Info().Msg("Initialization finished")
	return nil
}

func main() {

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	if os.Getenv("DEBUG") == "true" {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.TimestampFieldName = "timestamp"
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05.000", NoColor: false}
		fileWriter := &lumberjack.Logger{
			Filename:   "logfile.log",
			MaxSize:    100, // in megabytes
			MaxBackups: 3,   // number of old log files to keep
			MaxAge:     28,  // in days
		}
		multiWriter := zerolog.MultiLevelWriter(consoleWriter, zerolog.SyncWriter(fileWriter))
		log.Logger = log.Output(multiWriter)
	}

	log.Info().Msg("Initializing webserver..")
	e := echo.New()

	log.Info().Msg("Registering http handlers..")
	web.RegisterHandlers(e)
	go func() {
		log.Info().Msg("Starting webserver..")
		e.Logger.Fatal(e.Start(":8080"))
	}()

	wsServer := jrws.NewWebsocketServer("/ws", 8888)

	err := InitConbeeInverterAndProperties()

	if err == nil {
		startupStatus := CheckSystemStatus(properties, conbeeClient, inverter)
		if startupStatus.Status == models.InitStatusOk {
			monitoring = NewMonitoringController(conbeeClient, inverter, wsServer)
			monitoring.StartMonitoring(*properties)
		}
	}

	//var responseParams models.InitResponseParams
	wsServer.AddHandler("status", func(request models2.Request, conws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: status")
		status := CheckSystemStatus(properties, conbeeClient, inverter)
		log.Info().Int("status", status.Status).Msg(status.StatusMessage)
		if status.Status != models.InitStatusOk {
			return status, nil
		}
		if monitoring != nil {
			wsServer.WriteNotificationToAllMembers("monitoring", models.MonitoringEnabledParams{Enabled: monitoring.isRunning})
			_, err := monitoring.RequestDataAndSendWsNotification(*properties)
			if err != nil {
				log.Error().Stack().Err(err).Msg("Error requesting data and sending ws notification")
			}
		} else {
			wsServer.WriteNotificationToAllMembers("monitoring", models.MonitoringEnabledParams{Enabled: false})
		}
		return status, nil
	})

	wsServer.AddHandler("init", func(request models2.Request, ws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: init")
		status := authenticateConbeeClient(properties, conbeeClient)
		return status, nil
	})

	wsServer.AddHandler("getLights", func(request models2.Request, ws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: getLights")
		lights, restErrResp, err := conbeeClient.GetLights()
		if err != nil {
			return nil, err
		}
		if restErrResp != nil && restErrResp.Code != 200 {
			return restErrResp, nil
		}
		return lights, nil
	})

	wsServer.AddHandler("loginKostal", func(request models2.Request, concurrentWs *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: loginKostal")
		authParams := &models.AuthenticateParams{}
		err := jrws.CreateParamsObject(request.Params, authParams)
		if err != nil {
			return nil, err
		}

		properties.KostalAddress = authParams.HostAddress
		properties.KostalUsername = authParams.Username
		properties.KostalPassword = authParams.Password
		log.Info().Msgf("KostalAddress: %s, KostalUsername: %s", properties.KostalAddress, properties.KostalUsername)

		inverter = NewInverterClient(authParams.HostAddress, authParams.Password, properties.KostalType)
		err = inverter.Connect()
		if err != nil {
			log.Error().Err(err).Msg("Error connecting to inverter")
			return models.InitResponseParams{
				Status:        models.InitStatusKostalAuth,
				StatusMessage: err.Error(),
			}, nil

		}
		err = ini.SavePropertiesToFile("config.ini", properties.ToMap())
		if err != nil {
			return models.InitResponseParams{
				Status:        models.InitStatusError,
				StatusMessage: err.Error(),
			}, nil
		}

		status := CheckSystemStatus(properties, conbeeClient, inverter)
		IfStatusOk_InitAndStartMonitoring(status, wsServer)

		return status, nil

	})

	wsServer.AddHandler("authenticate",
		func(request models2.Request, concurrentWs *jrws.ConcurrentWebsocket) (interface{}, error) {
			log.Info().Msg("Handler: authenticate")
			authParams := &models.AuthenticateParams{}
			err := jrws.CreateParamsObject(request.Params, authParams)
			if err != nil {
				return nil, err
			}

			log.Info().Msgf("Username: %s", authParams.Username)
			properties.DeconzUsername = authParams.Username
			properties.DeconzPassword = authParams.Password
			conbeeClient.username = authParams.Username
			conbeeClient.password = authParams.Password

			if authParams.HostAddress != "" {
				properties.HostAddress = authParams.HostAddress
				conbeeClient.hostAddress = authParams.HostAddress
				conbeeClient.restClient.SetBaseURL("http://" + conbeeClient.hostAddress)
			}
			err = ini.SavePropertiesToFile("config.ini", properties.ToMap())
			if err != nil {
				return nil, err
			}

			status := authenticateConbeeClient(properties, conbeeClient)
			if status.Status != models.InitStatusOk {
				return status, nil
			}

			status = CheckSystemStatus(properties, conbeeClient, inverter)
			IfStatusOk_InitAndStartMonitoring(status, wsServer)
			return status, nil
		})

	wsServer.AddHandler("getProperties", func(request models2.Request, ws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: getProperties")
		return properties, nil
	})

	wsServer.AddHandler("startMonitoring", func(request models2.Request, ws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: startMonitoring")
		if monitoring == nil {
			monitoring = NewMonitoringController(conbeeClient, inverter, wsServer)
		}
		monitoring.StartMonitoring(*properties)
		return nil, nil
	})

	wsServer.AddHandler("stopMonitoring", func(request models2.Request, ws *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: stopMonitoring")
		monitoring.StopMonitoring()
		return nil, nil
	})

	wsServer.AddHandler("saveProperties", func(request models2.Request, concurrentWs *jrws.ConcurrentWebsocket) (interface{}, error) {
		log.Info().Msg("Handler: saveProperties")
		saveProps := &models.SavePropertiesParams{}
		err := jrws.CreateParamsObject(request.Params, saveProps)
		if err != nil {
			return nil, err
		}
		properties.PlugName = saveProps.PlugName
		properties.Threshold = saveProps.Threshold
		properties.PollDuration = saveProps.PollDuration
		err = ini.SavePropertiesToFile("config.ini", properties.ToMap())
		if err != nil {
			return nil, err
		}

		status := CheckSystemStatus(properties, conbeeClient, inverter)
		if status.Status != models.InitStatusOk {
			return status, nil
		}

		if monitoring == nil {
			monitoring = NewMonitoringController(conbeeClient, inverter, wsServer)
		}
		if monitoring.IsRunning() {
			log.Info().Msg("Restart monitoring")
			monitoring.StartMonitoring(*properties)
		} else {
			log.Info().Msg("Start monitoring")
		}

		return models.InitResponseParams{
			Status:        models.InitStatusOk,
			StatusMessage: "Everything is fine",
		}, nil
	})

	wsServer.AddHandler("switchLightOn", func(request models2.Request, concurrentWs *jrws.ConcurrentWebsocket) (interface{}, error) {
		switchLightParams := &models.SwitchLightParams{}
		err := jrws.CreateParamsObject(request.Params, switchLightParams)
		if err != nil {
			return models.InitResponseParams{
				Status:        models.InitStatusError,
				StatusMessage: err.Error(),
			}, nil
		}

		log.Info().Msg("Handler: switchLightOn")
		status := CheckSystemStatus(properties, conbeeClient, inverter)
		if status.Status != models.InitStatusOk {
			return status, nil
		}
		err = conbeeClient.SwitchOnLight(switchLightParams.LightId)
		if err != nil {
			return models.InitResponseParams{
				Status:        models.InitStatusError,
				StatusMessage: err.Error(),
			}, nil
		}
		if monitoring != nil {
			monitoring.RequestDataAndSendWsNotification(*properties)
		}
		return models.InitResponseParams{
			Status:        models.InitStatusOk,
			StatusMessage: "Everything is fine",
		}, nil
	})

	wsServer.AddHandler("switchLightOff", func(request models2.Request, concurrentWs *jrws.ConcurrentWebsocket) (interface{}, error) {
		switchLightParams := &models.SwitchLightParams{}
		err := jrws.CreateParamsObject(request.Params, switchLightParams)
		if err != nil {
			return models.InitResponseParams{
				Status:        models.InitStatusError,
				StatusMessage: err.Error(),
			}, nil
		}

		log.Info().Msg("Handler: switchLightOn")
		status := CheckSystemStatus(properties, conbeeClient, inverter)
		if status.Status != models.InitStatusOk {
			return status, nil
		}
		err = conbeeClient.SwitchOffLight(switchLightParams.LightId)
		if err != nil {
			return models.InitResponseParams{
				Status:        models.InitStatusError,
				StatusMessage: err.Error(),
			}, nil
		}
		if monitoring != nil {
			monitoring.RequestDataAndSendWsNotification(*properties)
		}
		return models.InitResponseParams{
			Status:        models.InitStatusOk,
			StatusMessage: "Everything is fine",
		}, nil
	})

	log.Info().Msg("Starting websocket server..")
	wsServer.StartListening()
}

func IfStatusOk_InitAndStartMonitoring(startupStatus models.InitResponseParams, wsServer *jrws.WebsocketServer) {
	if startupStatus.Status == models.InitStatusOk {
		if monitoring == nil {
			monitoring = NewMonitoringController(conbeeClient, inverter, wsServer)
		}
		_, err := monitoring.RequestDataAndSendWsNotification(*properties)
		if err != nil {
			log.Error().Stack().Err(err).Msg("Error requesting data and sending ws notification")
			return
		}
		monitoring.StartMonitoring(*properties)
	}
}

func CheckSystemStatus(properties *models.Properties, conbeeClient *ConbeeClient, inverter Inverter) models.InitResponseParams {
	log.Info().Msg("Check system status")

	if properties == nil {
		log.Info().Msg("Properties are nil")
		return models.InitResponseParams{
			Status:        models.InitStatusError,
			StatusMessage: "Properties are nil",
		}
	}

	if conbeeClient == nil {
		log.Info().Msg("Conbee client is nil")
		return models.InitResponseParams{
			Status:        models.InitStatusDeconzAuth,
			StatusMessage: "Conbee client is nil",
		}
	}

	valid, err := conbeeClient.CheckApiKey()
	if err != nil {
		log.Error().Err(err).Msg("Error while checking api key")
		return models.InitResponseParams{
			Status:        models.InitStatusDeconzAuth,
			StatusMessage: err.Error() + "Error while checking api key",
		}
	}
	if !valid {
		log.Info().Msg("Deconz Api key not valid")
		return models.InitResponseParams{
			Status:        models.InitStatusDeconzAuth,
			StatusMessage: "Deconz Api key not valid",
		}
	}
	log.Info().Msg("Deconz Api key is valid")

	if inverter == nil {
		log.Info().Msg("Inverter is nil")
		return models.InitResponseParams{
			Status:        models.InitStatusKostalAuth,
			StatusMessage: "Inverter is nil",
		}
	}

	if !inverter.IsConnected() {
		log.Info().Msg("Inverter not connected")
		return models.InitResponseParams{
			Status:        models.InitStatusKostalAuth,
			StatusMessage: "Inverter not connected",
		}
	}

	if properties.PlugName == "" {
		return models.InitResponseParams{
			Status:        models.InitStatusConfig,
			StatusMessage: "No socket name configured",
		}
	}
	if properties.PollDuration == 0 {
		return models.InitResponseParams{
			Status:        models.InitStatusConfig,
			StatusMessage: "No poll duration configured",
		}
	}
	if properties.Threshold == 10000 {
		return models.InitResponseParams{
			Status:        models.InitStatusConfig,
			StatusMessage: "No threshold configured",
		}
	}

	return models.InitResponseParams{
		Status:        models.InitStatusOk,
		StatusMessage: "Everything is fine",
	}
}
