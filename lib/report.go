package lib

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
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

type Section struct {
	Title      string            `json:"title"`
	Text       []string          `json:"text"`
	LocalHost  *ReportLocalHost  `json:"localhost"`
	RemoteHost *ReportRemoteHost `json:"remotehost"`
}

type ReportData struct {
	ReportID   ReportID  `dynamo:"report_id"`
	DataID     string    `dynamo:"data_id"`
	Data       []byte    `dynamo:"data"`
	TimeToLive time.Time `dynamo:"ttl"`
}

func NewReportData(reportID ReportID) *ReportData {
	data := ReportData{
		ReportID: reportID,
		DataID:   uuid.NewV4().String(),
	}

	return &data
}

func (x *ReportData) SetSection(section Section) {
	data, err := json.Marshal(&section)
	if err != nil {
		log.Println("Fail to marshal report section:", section)
	}

	x.Data = data
}

func (x *ReportData) Section() *Section {
	if len(x.Data) == 0 {
		return nil
	}

	var section Section
	err := json.Unmarshal(x.Data, &section)
	if err != nil {
		log.Println("Invalid report section data foramt", string(x.Data))
		return nil
	}

	return &section
}

func (x *ReportData) Submit(tableName, region string) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	x.TimeToLive = time.Now().Add(time.Second * 864000)

	err := table.Put(x).Run()
	if err != nil {
		return errors.Wrap(err, "Fail to put report data")
	}

	return nil
}

func FetchReportData(tableName, region string, reportID ReportID) ([]*Section, error) {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	dataList := []ReportData{}
	err := table.Get("report_id", reportID).All(&dataList)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report data")
	}

	sections := []*Section{}
	for _, data := range dataList {
		sections = append(sections, data.Section())
	}
	return sections, nil
}

type ReportResult struct {
	Severity string `json:"severity"`
}

type Report struct {
	ID       ReportID      `json:"report_id"`
	Alert    Alert         `json:"alert"`
	Sections []*Section    `json:"sections"`
	Result   *ReportResult `json:"result"`
}

func NewReport(reportID ReportID, alert *Alert) *Report {
	report := Report{
		ID:       reportID,
		Sections: []*Section{},
	}
	if alert != nil {
		report.Alert = *alert
	}

	return &report
}

func NewReportID() ReportID {
	return ReportID(uuid.NewV4().String())
}
