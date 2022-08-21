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
	Id          uint64    `json:'id'`
	Description string    `json:'description'`
	Start       time.Time `json:'start'`
	Stop        time.Time `json:'stop'`
	Duration    int64     `json:'duration'`
	Billable    bool      `json:'billable'`
	Wid         uint64    `json:'wid'`
	Pid         uint64    `json:'pid'`
	Tags        []string  `json:'tags'`
}

// Project struct defines the Project entry.
type Project struct {
	Data ProjectData `json:'data'`
}

// ProjectData struct defines the Project Data entry.
type ProjectData struct {
	Id        uint64    `json:'id'`
	Wid       uint64    `json:'wid'`
	Cid       uint64    `json:'cid'`
	Name      string    `json:'name'`
	Billable  bool      `json:'billable'`
	IsPrivate bool      `json:'is_private'`
	Active    bool      `json:'active'`
	At        time.Time `json:'at'`
	Template  bool      `json:'template'`
	Color     string    `json:'color'`
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
	url := "https://api.track.toggl.com/api/v8/projects/" + strconv.FormatUint(projectId, 10)
	resp, err := togglClient.executeHttpGet(url)
	if err != nil {
		return ProjectData{}, err
	}
	defer resp.Body.Close()

	var project Project
	unmrshalErr := json.NewDecoder(resp.Body).Decode(&project)
	if unmrshalErr != nil {
		return ProjectData{}, unmrshalErr
	}
	return project.Data, nil
}

// GetRange retrieves the Toggl Time entries between the startTime and endTime time range.
func (togglClient *TogglClient) GetRange(startTime time.Time, endTime time.Time) ([]TogglTimeEntry, error) {
	url := "https://api.track.toggl.com/api/v8/time_entries?start_date=" + startTime.Format("2006-01-02") + "T00%3A00%3A00%2B00%3A00&end_date=" + endTime.Format("2006-01-02") + "T23%3A59%3A59%2B00%3A00"
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
