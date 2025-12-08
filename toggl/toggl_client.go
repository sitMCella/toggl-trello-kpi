package toggl

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// TogglClient implements the Toggl Client interface.
type TogglClient struct {
	logger        *zap.Logger
	configuration configuration.TogglConfiguration
	projectsData  map[uint64]ProjectData
}

// TimeEntry struct defines the Toggl Time entry.
type TimeEntry struct {
	Id              uint64    `json:'id'`
	Description     string    `json:'description'`
	Start           time.Time `json:'start'`
	Stop            time.Time `json:'stop'`
	Duration        int64     `json:'duration'`
	Billable        bool      `json:'billable'`
	WorkspaceId     uint64    `json:'workspace_id'`
	Wid             uint64    `json:'wid'`
	ProjectId       uint64    `json:'project_id'`
	Pid             uint64    `json:'pid'`
	TaskId          uint64    `json:'task_id'`
	Tags            []string  `json:'tags'`
	TagIds          []uint64  `json:'tag_ids'`
	Duronly         bool      `json:'duronly'`
	At              time.Time `json:'at'`
	ServerDeletedAt time.Time `json:'server_deleted_at'`
	UserId          uint64    `json:'user_id'`
	Uid             uint64    `json:'uid'`
}

// Project struct defines the Project entry.
type Project struct {
	Data ProjectData `json:'data'`
}

// ProjectData struct defines the Project Data entry.
type ProjectData struct {
	Id                  uint64    `json:'id'`
	WorkspaceId         uint64    `json:'workspace_id'`
	Wid                 uint64    `json:'wid'`
	ClientId            uint64    `json:'client_id'`
	Cid                 uint64    `json:'cid'`
	ClientName          string    `json:'client_name'`
	Name                string    `json:'name'`
	Billable            bool      `json:'billable'`
	IsPrivate           bool      `json:'is_private'`
	Active              bool      `json:'active'`
	At                  time.Time `json:'at'`
	CreatedAt           time.Time `json:'created_at'`
	ServerDeletedAt     time.Time `json:'server_deleted_at'`
	Template            bool      `json:'template'`
	TemplateId          uint64    `json:'template_id'`
	Color               string    `json:'color'`
	RecurringParameters string    `json:'recurring_parameters'`
	FixedFee            string    `json:'fixed_fee'`
	ActualHours         uint64    `json:'actual_hours'`
	ActualSeconds       uint64    `json:'actual_seconds'`
	TotalCount          uint64    `json:'total_count'`
	CanTrackTime        bool      `json:'can_track_time'`
	StartDate           time.Time `json:'start_date'`
	Status              string    `json:'status'`
}

// NewTogglClient creates a new TogglClient.
func NewTogglClient(config configuration.Configuration, logger *zap.Logger) *TogglClient {
	return &TogglClient{
		logger:        logger,
		configuration: config.TogglConfiguration,
		projectsData:  make(map[uint64]ProjectData),
	}
}

// GetProjectData retrieves the Project Data from a Project ID.
func (togglClient *TogglClient) GetProjectData(projectId uint64) (ProjectData, error) {
	url := "https://api.track.toggl.com/api/v9/workspaces/" + togglClient.configuration.WorkspaceId + "/projects/" + strconv.FormatUint(projectId, 10)
	resp, err := togglClient.executeHttpGet(url)
	if err != nil {
		return ProjectData{}, err
	}
	defer resp.Body.Close()

	var projectData ProjectData
	unmrshalErr := json.NewDecoder(resp.Body).Decode(&projectData)
	if unmrshalErr != nil {
		return ProjectData{}, unmrshalErr
	}
	return projectData, nil
}

// GetRange retrieves the Toggl Time entries between the startTime and endTime time range.
func (togglClient *TogglClient) GetRange(startTime time.Time, endTime time.Time) ([]TogglTimeEntry, error) {
	url := "https://api.track.toggl.com/api/v9/me/time_entries?start_date=" + startTime.Format("2006-01-02") + "&end_date=" + endTime.Format("2006-01-02")
	resp, err := togglClient.executeHttpGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var timeEntries []TimeEntry
	unmrshalErr := json.NewDecoder(resp.Body).Decode(&timeEntries)
	if unmrshalErr != nil {
		return nil, unmrshalErr
	}

	var togglTimeEntries = make([]TogglTimeEntry, len(timeEntries))
	for i, timeEntry := range timeEntries {
		projectName, err := togglClient.getProjectName(timeEntry)
		if err != nil {
			return nil, err
		}
		togglTimeEntry := TogglTimeEntry{
			Id:             timeEntry.Id,
			Description:    timeEntry.Description,
			Start:          timeEntry.Start,
			Stop:           timeEntry.Stop,
			Duration:       timeEntry.Duration,
			Billable:       timeEntry.Billable,
			Workspace_id:   timeEntry.Wid,
			Project_id:     timeEntry.Pid,
			Project_name:   projectName,
			Tags:           timeEntry.Tags,
			Trello_card_id: "",
		}
		togglTimeEntries[i] = togglTimeEntry
	}
	return togglTimeEntries, nil
}

// getAuthorizationHeader configures the Authorization Http Header from the Toggl API Token.
func (togglClient *TogglClient) getAuthorizationHeader() string {
	apiToken := togglClient.configuration.ApiToken + ":api_token"
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(apiToken))
}

// executeHttpGet executes the Http call from a URL and returns the response object.
func (togglClient *TogglClient) executeHttpGet(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", togglClient.getAuthorizationHeader())
	return client.Do(req)
}

// getProjectName retrieves the Project Name from the Project ID in the TimeEntry object.
func (togglClient *TogglClient) getProjectName(timeEntry TimeEntry) (string, error) {
	if projectData, found := togglClient.projectsData[timeEntry.Pid]; found {
		return projectData.Name, nil
	} else {
		projectData, err := togglClient.GetProjectData(timeEntry.Pid)
		if err != nil {
			return "", err
		}
		togglClient.projectsData[timeEntry.Pid] = projectData
		return projectData.Name, nil
	}
}
