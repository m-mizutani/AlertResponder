package lib

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type ReportID string

type ReportMalware struct {
	Name string
}

type ReportServiceUsage struct {
	ServiceName string    `json:"service_name"`
	Principal   string    `json:"principal"`
	LastSeen    time.Time `json:"last_seen"`
}

type ReportLocalHost struct {
	UserName     []string              `json:"username"`
	OS           []string              `json:"os"`
	IPAddr       []string              `json:"ipaddr"`
	Country      []string              `json:"country"`
	ServiceUsage []*ReportServiceUsage `json:"service_usage"`
}

type ReportRemoteHost struct {
	IPAddr         []string `json:"ipaddr"`
	Domain         []string `json:"domain"`
	Country        []string `json:"country"`
	RelatedMalware []string `json:"related_malware"`
	RelatedDomains []string `json:"related_domains"`
}

type ReportData struct {
	Title      string            `json:"title"`
	Text       []string          `json:"text"`
	LocalHost  *ReportLocalHost  `json:"localhost"`
	RemoteHost *ReportRemoteHost `json:"remotehost"`
}

type Report struct {
	ID    ReportID      `json:"report_id"`
	Alert Alert         `json:"alert"`
	Data  []*ReportData `json:"data"`
}

func NewReport(reportID ReportID, alert *Alert) *Report {
	report := Report{
		ID:   reportID,
		Data: []*ReportData{},
	}
	if alert != nil {
		report.Alert = *alert
	}

	return &report
}

func NewReportID() ReportID {
	return ReportID(uuid.NewV4().String())
}
