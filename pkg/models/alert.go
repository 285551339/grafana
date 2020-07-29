package models

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
)

type AlertStateType string
type AlertSeverityType string
type NoDataOption string
type ExecutionErrorOption string

const (
	AlertStateNoData   AlertStateType = "no_data"
	AlertStatePaused   AlertStateType = "paused"
	AlertStateAlerting AlertStateType = "alerting"
	AlertStateOK       AlertStateType = "ok"
	AlertStatePending  AlertStateType = "pending"
	AlertStateUnknown  AlertStateType = "unknown"
)

const (
	NoDataSetOK       NoDataOption = "ok"
	NoDataSetNoData   NoDataOption = "no_data"
	NoDataKeepState   NoDataOption = "keep_state"
	NoDataSetAlerting NoDataOption = "alerting"
)

const (
	ExecutionErrorSetAlerting ExecutionErrorOption = "alerting"
	ExecutionErrorKeepState   ExecutionErrorOption = "keep_state"
)

var (
	ErrCannotChangeStateOnPausedAlert = fmt.Errorf("Cannot change state on pause alert")
	ErrRequiresNewState               = fmt.Errorf("update alert state requires a new state")
)

func (s AlertStateType) IsValid() bool {
	return s == AlertStateOK ||
		s == AlertStateNoData ||
		s == AlertStatePaused ||
		s == AlertStatePending ||
		s == AlertStateAlerting ||
		s == AlertStateUnknown
}

func (s NoDataOption) IsValid() bool {
	return s == NoDataSetNoData || s == NoDataSetAlerting || s == NoDataKeepState || s == NoDataSetOK
}

func (s NoDataOption) ToAlertState() AlertStateType {
	return AlertStateType(s)
}

func (s ExecutionErrorOption) IsValid() bool {
	return s == ExecutionErrorSetAlerting || s == ExecutionErrorKeepState
}

func (s ExecutionErrorOption) ToAlertState() AlertStateType {
	return AlertStateType(s)
}

type Alert struct {
	Id             int64
	Version        int64
	OrgId          int64
	DashboardId    int64
	PanelId        int64
	Name           string
	Message        string
	Severity       string //Unused
	State          AlertStateType
	Handler        int64 //Unused
	Silenced       bool
	ExecutionError string
	Frequency      int64
	For            time.Duration

	EvalData     *simplejson.Json
	NewStateDate time.Time
	StateChanges int64

	Created time.Time
	Updated time.Time

	Settings *simplejson.Json
}

func (alert *Alert) ValidToSave() bool {
	return alert.DashboardId != 0 && alert.OrgId != 0 && alert.PanelId != 0
}

func (alert *Alert) ShouldUpdateState(newState AlertStateType) bool {
	return alert.State != newState
}

func (this *Alert) ContainsUpdates(other *Alert) bool {
	result := false
	result = result || this.Name != other.Name
	result = result || this.Message != other.Message

	if this.Settings != nil && other.Settings != nil {
		json1, err1 := this.Settings.Encode()
		json2, err2 := other.Settings.Encode()

		if err1 != nil || err2 != nil {
			return false
		}

		result = result || string(json1) != string(json2)
	}

	//don't compare .State! That would be insane.
	return result
}

func (alert *Alert) GetTagsFromSettings() []*Tag {
	tags := []*Tag{}
	if alert.Settings != nil {
		if data, ok := alert.Settings.CheckGet("alertRuleTags"); ok {
			for tagNameString, tagValue := range data.MustMap() {
				// MustMap() already guarantees the return of a `map[string]interface{}`.
				// Therefore we only need to verify that tagValue is a String.
				tagValueString := simplejson.NewFromAny(tagValue).MustString()
				tags = append(tags, &Tag{Key: tagNameString, Value: tagValueString})
			}
		}
	}
	return tags
}

func (alert *Alert) GetNotificationsFromSettings() []string {
	notifications := make([]string, 0)
	if alert.Settings != nil {
		if data, ok := alert.Settings.CheckGet("notifications"); ok {
			data := data.MustArray()
			for _, m := range data {
				mm := simplejson.NewFromAny(m).MustMap()
				notifications = append(notifications, mm["uid"].(string))
			}
		}
	}
	return notifications
}

type AlertingClusterInfo struct {
	ServerId       string
	ClusterSize    int
	UptimePosition int
}

type HeartBeat struct {
	Id       int64
	ServerId string
	Updated  time.Time
	Created  time.Time
}

type HeartBeatCommand struct {
	ServerId string
	Result   AlertingClusterInfo
}

type SaveAlertsCommand struct {
	DashboardId int64
	UserId      int64
	OrgId       int64

	Alerts []*Alert
}

type PauseAlertCommand struct {
	OrgId       int64
	AlertIds    []int64
	ResultCount int64
	Paused      bool
}

type PauseAllAlertCommand struct {
	ResultCount int64
	Paused      bool
}

type SetAlertStateCommand struct {
	AlertId  int64
	OrgId    int64
	State    AlertStateType
	Error    string
	EvalData *simplejson.Json

	Result Alert
}

//Queries
type GetAlertsQuery struct {
	OrgId                   int64
	State                   []string
	DashboardIDs            []int64
	PanelId                 int64
	Limit                   int64
	Query                   string
	User                    *SignedInUser
	StandaloneAlertsEnabled bool

	Result []*AlertListItemDTO
}

type GetAllAlertsQuery struct {
	Result []*Alert
}

type GetAlertByIdQuery struct {
	Id int64

	Result *Alert
}

type GetAlertStatesForDashboardQuery struct {
	OrgId       int64
	DashboardId int64

	Result []*AlertStateInfoDTO
}

type AlertListItemDTO struct {
	Id             int64            `json:"id"`
	DashboardId    int64            `json:"dashboardId"`
	DashboardUid   string           `json:"dashboardUid"`
	DashboardSlug  string           `json:"dashboardSlug"`
	PanelId        int64            `json:"panelId"`
	Name           string           `json:"name"`
	State          AlertStateType   `json:"state"`
	NewStateDate   time.Time        `json:"newStateDate"`
	EvalDate       time.Time        `json:"evalDate"`
	EvalData       *simplejson.Json `json:"evalData"`
	ExecutionError string           `json:"executionError"`
	Url            string           `json:"url"`
}

type AlertStateInfoDTO struct {
	Id           int64          `json:"id"`
	DashboardId  int64          `json:"dashboardId"`
	PanelId      int64          `json:"panelId"`
	State        AlertStateType `json:"state"`
	NewStateDate time.Time      `json:"newStateDate"`
}

// "Internal" commands

type UpdateDashboardAlertsCommand struct {
	OrgId     int64
	Dashboard *Dashboard
	User      *SignedInUser
}

type ValidateDashboardAlertsCommand struct {
	UserId    int64
	OrgId     int64
	Dashboard *Dashboard
	User      *SignedInUser
}

type DeleteAlertCommand struct {
	Id int64
}

type alertCondition struct {
	Evaluator struct {
		Params []int64 `json:"params"`
		Type   string  `json:"type"`
	} `json:"evaluator"`
	Operator struct {
		Type string `json:"type"`
	} `json:"operator"`
	Query struct {
		DatasourceID int64 `json:"datasourceId"`
		Model        struct {
			data interface{}
		} `json:"model"`
		Params []string `json:"params"`
	} `json:"query"`
	Reducer struct {
		Params []string `json:"params"`
		Type   string   `json:"type"`
	} `json:"reducer"`
	Type string `json:"type"`
}

type CreateAlertCommand struct {
	AlertRuleTags       map[string]string `json:"alertRuleTags"`
	Conditions          []*alertCondition `json:"conditions"`
	ExecutionErrorState string            `json:"executionErrorState"`
	For                 string            `json:"for"`
	Frequency           int64             `json:"frequency"`
	Handler             int64             `json:"handler"`
	Name                string            `json:"name"`
	NoDataState         NoDataOption      `json:"noDataState"`
	Notifications       []struct {
		UID string `json:"uid"`
	} `json:"notifications"`

	OrgId  int64 `json:"-"`
	Result *Alert
}

type UpdateAlertCommand struct {
	OrgID               int64             `json:"-"`
	ID                  int64             `json:"id"`
	AlertRuleTags       map[string]string `json:"alertRuleTags"`
	Conditions          []*alertCondition `json:"conditions"`
	ExecutionErrorState string            `json:"executionErrorState"`
	For                 string            `json:"for"`
	Frequency           int64             `json:"frequency"`
	Handler             int64             `json:"handler"`
	Name                string            `json:"name"`
	NoDataState         NoDataOption      `json:"noDataState"`
	Notifications       []struct {
		UID string `json:"uid"`
	} `json:"notifications"`

	Result *Alert
}
