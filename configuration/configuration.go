package configuration

import (
	"fmt"
	"os"

	"github.com/ory/viper"
)

// Configuration struct defines the configuration properties.
type Configuration struct {
	ApplicationConfiguration
	TogglConfiguration
	TrelloConfiguration
	DBConfiguration
	GrafanaConfiguration
}

// ApplicationConfiguration struct defines the application configuration properties.
type ApplicationConfiguration struct {
	LogLevel string
}

// TogglConfiguration struct defines the Toggl configuration properties.
type TogglConfiguration struct {
	ApiToken string
}

// TrelloConfiguration struct define the Trello configuration properties.
type TrelloConfiguration struct {
	AppKey             string
	ApiToken           string
	BoardId            string
	LabelProjectColor  []string
	LabelCustomerColor []string
	LabelTeamColor     []string
	LabelCardTypeColor []string
}

// DBConfiguration struct defines the database configuration properties.
type DBConfiguration struct {
	Host                           string
	Port                           int
	Name                           string
	Username                       string
	Password                       string
	MaxOpenConnections             int
	MaxIdleConnections             int
	ConnectionMaxLifeTimeInMinutes int
}

// GrafanaConfiguration struc defines the Grafana configuration properties.
type GrafanaConfiguration struct {
	Year       string
	StartMonth string
	EndMonth   string
}

// FileNotExistsError defines the file not exists error.
type FileNotExistsError struct {
	SettingsFilePath string
}

func (err *FileNotExistsError) Error() string {
	return fmt.Sprintf("The Configuration settings file %s does not exist.", err.SettingsFilePath)
}

// ConfigurationSettingsError defines the configuration settings error.
type ConfigurationSettingsError struct {
	err error
}

func (err *ConfigurationSettingsError) Error() string {
	return fmt.Sprintf("Configuration settings error: %+v", err.err)
}

// NewConfiguration creates a new Configuration from the "settings.yml" file.
func NewConfiguration() (Configuration, error) {
	settingsFilePath := "./configuration/settings.yml"
	err := verifyConfigurationSettingsFileExists(settingsFilePath)
	if err != nil {
		return Configuration{}, err
	}
	viper.SetConfigFile(settingsFilePath)
	err = viper.ReadInConfig()
	if err != nil {
		return Configuration{}, &ConfigurationSettingsError{err: err}
	}
	viper.AutomaticEnv()
	applicationConfiguration := newApplicationConfiguration(viper.GetViper())
	togglConfiguration := newTogglConfiguration(viper.GetViper())
	trelloConfiguration := newTrelloConfiguration(viper.GetViper())
	dbConfiguration := newDatabaseConfiguration(viper.GetViper())
	grafanaConfiguration := newGrafanaConfiguration(viper.GetViper())
	return Configuration{
		ApplicationConfiguration: applicationConfiguration,
		TogglConfiguration:       togglConfiguration,
		TrelloConfiguration:      trelloConfiguration,
		DBConfiguration:          dbConfiguration,
		GrafanaConfiguration:     grafanaConfiguration,
	}, nil
}

func verifyConfigurationSettingsFileExists(settingsFilePath string) error {
	if _, err := os.Stat(settingsFilePath); err != nil {
		return &FileNotExistsError{SettingsFilePath: settingsFilePath}
	}
	return nil
}

func newApplicationConfiguration(viper *viper.Viper) ApplicationConfiguration {
	applicationLogLevel := viper.GetString("APPLICATION_LOG_LEVEL")
	return ApplicationConfiguration{
		LogLevel: applicationLogLevel,
	}
}

func newTogglConfiguration(viper *viper.Viper) TogglConfiguration {
	apiToken := viper.GetString("TOGGL_API_TOKEN")
	return TogglConfiguration{
		ApiToken: apiToken,
	}
}

func newTrelloConfiguration(viper *viper.Viper) TrelloConfiguration {
	appKey := viper.GetString("TRELLO_APP_KEY")
	apiToken := viper.GetString("TRELLO_API_TOKEN")
	boardId := viper.GetString("TRELLO_BOARD_ID")
	labelProjectColor := viper.GetStringSlice("TRELLO_LABEL_PROJECT_COLOR")
	labelCustomerColor := viper.GetStringSlice("TRELLO_LABEL_CUSTOMER_COLOR")
	labelTeamColor := viper.GetStringSlice("TRELLO_LABEL_TEAM_COLOR")
	labelCardTypeColor := viper.GetStringSlice("TRELLO_LABEL_CARD_TYPE_COLOR")
	return TrelloConfiguration{
		AppKey:             appKey,
		ApiToken:           apiToken,
		BoardId:            boardId,
		LabelProjectColor:  labelProjectColor,
		LabelCustomerColor: labelCustomerColor,
		LabelTeamColor:     labelTeamColor,
		LabelCardTypeColor: labelCardTypeColor,
	}
}

func newDatabaseConfiguration(viper *viper.Viper) DBConfiguration {
	databaseHost := viper.GetString("DATABASE_HOST")
	databasePort := viper.GetInt("DATABASE_PORT")
	databaseName := viper.GetString("DATABASE_NAME")
	databaseUsername := viper.GetString("DATABASE_USERNAME")
	databasePassword := viper.GetString("DATABASE_PASSWORD")
	maxOpenConnections := viper.GetInt("DATABASE_MAX_OPEN_CONNECTIONS")
	maxIdleConnections := viper.GetInt("DATABASE_MAX_IDLE_CONNECTIONS")
	connectionMaxLifeTimeInMinutes := viper.GetInt("DATABASE_MAX_LIFETIME_IN_MINUTES")
	return DBConfiguration{
		Host:                           databaseHost,
		Port:                           databasePort,
		Name:                           databaseName,
		Username:                       databaseUsername,
		Password:                       databasePassword,
		MaxOpenConnections:             maxOpenConnections,
		MaxIdleConnections:             maxIdleConnections,
		ConnectionMaxLifeTimeInMinutes: connectionMaxLifeTimeInMinutes,
	}
}

func newGrafanaConfiguration(viper *viper.Viper) GrafanaConfiguration {
	grafanaYear := viper.GetString("GRAFANA_YEAR")
	grafanaStartMonth := viper.GetString("GRAFANA_START_MONTH")
	grafanaEndMonth := viper.GetString("GRAFANA_END_MONTH")
	return GrafanaConfiguration{
		Year:       grafanaYear,
		StartMonth: grafanaStartMonth,
		EndMonth:   grafanaEndMonth,
	}
}
