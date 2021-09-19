package toggl

import (
	"time"

	"github.com/dougEfresh/gtoggl"
	"go.uber.org/zap"
)

// TogglClient implements the Toggl Client interface.
type TogglClient struct {
	logger      *zap.Logger
	togglClient *gtoggl.TogglClient
}

// NewTogglClient creates a new TogglClient.
func NewTogglClient(logger *zap.Logger, togglClient *gtoggl.TogglClient) *TogglClient {
	return &TogglClient{
		logger:      logger,
		togglClient: togglClient,
	}
}

// GetRange retrieves the Toggl entries between the startTime and endTime time range.
func (togglClient *TogglClient) GetRange(startTime time.Time, endTime time.Time) ([]TogglTimeEntry, error) {
	timeEntries, err := togglClient.togglClient.TimeentryClient.GetRange(startTime, endTime)
	if err != nil {
		return nil, err
	}
	var togglTimeEntries = make([]TogglTimeEntry, len(timeEntries))
	for i, timeEntry := range timeEntries {
		togglTimeEntry := TogglTimeEntry{
			Id:             timeEntry.Id,
			Description:    timeEntry.Description,
			Start:          timeEntry.Start,
			Stop:           timeEntry.Stop,
			Duration:       timeEntry.Duration,
			Billable:       timeEntry.Billable,
			Workspace_id:   timeEntry.Wid,
			Project_id:     timeEntry.Pid,
			Tags:           timeEntry.Tags,
			Trello_card_id: "",
		}
		togglTimeEntries[i] = togglTimeEntry
	}
	return togglTimeEntries, nil
}
