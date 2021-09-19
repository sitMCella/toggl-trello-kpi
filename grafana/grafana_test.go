package grafana

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/sitMCella/toggl-trello-kpi/application_errors"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"go.uber.org/zap"
)

func TestGrafanaDashboardCreateThrowsErrorOnNilLogger(t *testing.T) {
	config := configuration.Configuration{
		GrafanaConfiguration: configuration.GrafanaConfiguration{
			Year:       "2021",
			StartMonth: "02",
			EndMonth:   "03",
		},
	}

	_, err := NewGrafanaDashboard(config, nil)

	if err == nil {
		t.Fatalf("Expect an error while creating GrafanaDashboard with nil logger.")
	}
	switch err.(type) {
	case *application_errors.NilParameterError:
		return
	default:
		t.Errorf("Expect a NilParameterError while creating GrafanaDashboard with nil logger.")
	}
}

func TestCreateGrafanaDashboard(t *testing.T) {
	config := configuration.Configuration{
		GrafanaConfiguration: configuration.GrafanaConfiguration{
			Year:       "2021",
			StartMonth: "02",
			EndMonth:   "03",
		},
	}
	logger, err := getLogger()
	if err != nil {
		t.Errorf("Couldn't initialize logger: %v", err)
	}
	defer logger.Sync()
	grafanaDashboard, nil := NewGrafanaDashboard(config, logger)
	if err != nil {
		t.Fatalf("Error creating GrafanaDashboard: %v", err)
	}
	dashboardFilePath := "/tmp/grafana_dashboard_test.json"
	grafanaDashboard.dashboardFilePath = dashboardFilePath

	err = grafanaDashboard.CreateDashboard()
	if err != nil {
		t.Fatalf("Error creating Grafana Dashboard: %v", err)
	}

	grafanaDashboardData, err := ioutil.ReadFile(dashboardFilePath)
	if err != nil {
		t.Fatalf("Error retrieving Grafana Dashboard data: %v", err)
	}
	defer func() {
		oserr := os.Remove(dashboardFilePath)
		if oserr != nil {
			t.Fatalf("Error while removing the test file: %v", oserr)
		}
	}()
	expectedGrafanaDashboardData, err := ioutil.ReadFile("./dashboard_test.json")
	if err != nil {
		t.Errorf("Error retrieving expected Grafana Dashboard data: %v", err)
	}
	assert.Equal(t, expectedGrafanaDashboardData, grafanaDashboardData, "Expected same Grafana Dashboard")
}

func getLogger() (*zap.Logger, error) {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.FatalLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return zapCfg.Build()
}
