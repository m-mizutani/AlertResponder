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

type Report struct {
	ID     ReportID      `json:"report_id"`
	Alert  Alert         `json:"alert"`
	Pages  []*ReportPage `json:"pages"`
	Result *ReportResult `json:"result"`
}

type ReportPage struct {
	Title      string            `json:"title"`
	Text       []string          `json:"text"`
	LocalHost  *ReportLocalHost  `json:"localhost"`
	RemoteHost *ReportRemoteHost `json:"remotehost"`
}

type ReportResult struct {
	Severity string `json:"severity"`
}

type ReportMalware struct {
	SHA256    string              `json:"sha256"`
	Timestamp time.Time           `json:"timestamp"`
	Scans     []ReportMalwareScan `json:"scans"`
	Relation  string              `json:"relation"`
}

type ReportMalwareScan struct {
	Vendor   string `json:"vendor"`
	Name     string `json:"name"`
	Positive bool   `json:"positive"`
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
	IPAddr         []string        `json:"ipaddr"`
	Domain         []string        `json:"domain"`
	Country        []string        `json:"country"`
	RelatedMalware []ReportMalware `json:"related_malware"`
	RelatedDomains []string        `json:"related_domains"`
}

type ReportComponent struct {
	ReportID   ReportID  `dynamo:"report_id"`
	DataID     string    `dynamo:"data_id"`
	Data       []byte    `dynamo:"data"`
	TimeToLive time.Time `dynamo:"ttl"`
}

// NewReportComponent is a constructor of ReportComponent
func NewReportComponent(reportID ReportID) *ReportComponent {
	data := ReportComponent{
		ReportID: reportID,
		DataID:   uuid.NewV4().String(),
	}

	return &data
}

// Setpage sets page data with serialization.
func (x *ReportComponent) SetPage(page ReportPage) {
	data, err := json.Marshal(&page)
	if err != nil {
		log.Println("Fail to marshal report page:", page)
	}

	x.Data = data
}

// page returns deserialized page structure
func (x *ReportComponent) Page() *ReportPage {
	if len(x.Data) == 0 {
		return nil
	}

	var page ReportPage
	err := json.Unmarshal(x.Data, &page)
	if err != nil {
		log.Println("Invalid report page data foramt", string(x.Data))
		return nil
	}

	return &page
}

func (x *ReportComponent) Submit(tableName, region string) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	x.TimeToLive = time.Now().Add(time.Second * 864000)

	err := table.Put(x).Run()
	if err != nil {
		return errors.Wrap(err, "Fail to put report data")
	}

	return nil
}

func FetchReportPages(tableName, region string, reportID ReportID) ([]*ReportPage, error) {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	dataList := []ReportComponent{}
	err := table.Get("report_id", reportID).All(&dataList)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report data")
	}

	pages := []*ReportPage{}
	for _, data := range dataList {
		pages = append(pages, data.Page())
	}
	return pages, nil
}

func NewReport(reportID ReportID, alert *Alert) *Report {
	report := Report{
		ID:    reportID,
		Pages: []*ReportPage{},
	}
	if alert != nil {
		report.Alert = *alert
	}

	return &report
}

func NewReportID() ReportID {
	return ReportID(uuid.NewV4().String())
}
