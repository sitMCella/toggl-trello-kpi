package grafana

import (
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"go.uber.org/zap"
)

// GrafanaDashboardParametersParseError struct defines the Grafana Dashboard parameters parse error.
type GrafanaDashboardParametersParseError struct {
	parameterName string
}

func (err *GrafanaDashboardParametersParseError) Error() string {
	return fmt.Sprintf("Error parsing the Grafana Dashboard parameter %s.", err.parameterName)
}

// GrafanaDashboardError struct defines the Grafana Dashboard error.
type GrafanaDashboardError struct {
	errorMessage string
}

func (err *GrafanaDashboardError) Error() string {
	return fmt.Sprintf(err.errorMessage)
}

// GrafanaDashboard struct defines the Grafana Dashboard service.
type GrafanaDashboard struct {
	logger            *zap.Logger
	configuration     configuration.GrafanaConfiguration
	dashboardFilePath string
}

// GrafanaDashboardTemplateParameters struct defines the Grafana Dashboard template parameters.
type GrafanaDashboardTemplateParameters struct {
	StartTime string
	EndTime   string
}

// NewGrafanaDashboard creates a new GrafanaDashboard.
func NewGrafanaDashboard(config configuration.Configuration, logger *zap.Logger) (*GrafanaDashboard, error) {
	if logger == nil {
		return nil, &application_errors.NilParameterError{ParameterName: "logger"}
	}
	return &GrafanaDashboard{
		logger:            logger,
		configuration:     config.GrafanaConfiguration,
		dashboardFilePath: "./grafana/dashboard.json",
	}, nil
}

// CreateDashboard creates a new Grafana Dashboard.
func (grafanaDashboard *GrafanaDashboard) CreateDashboard() (err error) {
	grafanaDashboardTemplateParameters, err := grafanaDashboard.getDashboardTemplateParameters()
	if err != nil {
		return
	}
	box := packr.New("dashboard", "./grafana")
	dashboardTemplate, err := box.Find("dashboard_template.json")
	if err != nil {
		return
	}
	t := template.Must(template.New("manifest").Parse(string(dashboardTemplate)))
	dashboardFile, err := os.Create(grafanaDashboard.dashboardFilePath)
	if err != nil {
		return
	}
	defer func() {
		fileerr := dashboardFile.Close()
		if err == nil {
			err = fileerr
		}
	}()
	if err = t.Execute(dashboardFile, grafanaDashboardTemplateParameters); err != nil {
		return
	}
	return
}

func (grafanaDashboard *GrafanaDashboard) getDashboardTemplateParameters() (GrafanaDashboardTemplateParameters, error) {
	year, err := strconv.Atoi(grafanaDashboard.configuration.Year)
	if err != nil {
		return GrafanaDashboardTemplateParameters{}, &GrafanaDashboardParametersParseError{parameterName: "Year"}
	}
	startMonth, err := strconv.Atoi(grafanaDashboard.configuration.StartMonth)
	if err != nil {
		return GrafanaDashboardTemplateParameters{}, &GrafanaDashboardParametersParseError{parameterName: "StartMonth"}
	}
	endMonth, err := strconv.Atoi(grafanaDashboard.configuration.EndMonth)
	if err != nil {
		return GrafanaDashboardTemplateParameters{}, &GrafanaDashboardParametersParseError{parameterName: "EndMonth"}
	}
	if endMonth < startMonth {
		return GrafanaDashboardTemplateParameters{}, &GrafanaDashboardError{errorMessage: "The EndMonth parameter must be greter than or equal to the StartMonth parameter"}
	}
	startTime := fmt.Sprintf("%d-%02d-01T00:00:00.000Z", year, startMonth)
	endMonthBeginning := time.Date(year, time.Month(endMonth), 01, 0, 0, 0, 0, time.UTC)
	endMonthEnding := endMonthBeginning.AddDate(0, 1, -1)
	endTime := fmt.Sprintf("%d-%02d-%02dT23:59:59.999Z", year, endMonth, endMonthEnding.Day())
	grafanaDashboardTemplateParameters := GrafanaDashboardTemplateParameters{
		StartTime: startTime,
		EndTime:   endTime,
	}
	return grafanaDashboardTemplateParameters, nil
}
