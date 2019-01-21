package lib

import (
	"encoding/json"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type ReportID string
type ReportStatus string
type ReportSeverity string

type Report struct {
	ID      ReportID      `json:"report_id"`
	Alert   Alert         `json:"alert"`
	Content ReportContent `json:"content"`
	Result  ReportResult  `json:"result"`
	Status  ReportStatus  `json:"status"`
	// Status must be "new" or "published".
	//
	// new: This status means that the report is issued by Receptor.
	//      No inspect information
	// published: When publisher receives report with result, report status
	//            is "published".
	//
}

// IsNew and IsPublished returns status of the report
func (x *Report) IsNew() bool       { return x.Status == StatusNew }
func (x *Report) IsPublished() bool { return x.Status == StatusPublished }

const (
	StatusNew       ReportStatus = "new"
	StatusOngoing   ReportStatus = "ongoing"
	StatusPublished ReportStatus = "published"
)

type ReportContent struct {
	OpponentHosts map[string]ReportOpponentHost `json:"opponent_hosts"`
	AlliedHosts   map[string]ReportAlliedHost   `json:"allied_hosts"`
	SubjectUsers  map[string]ReportURL          `json:"subject_users"`
}

func newReportContent() ReportContent {
	return ReportContent{
		OpponentHosts: map[string]ReportOpponentHost{},
		AlliedHosts:   map[string]ReportAlliedHost{},
		SubjectUsers:  map[string]ReportURL{},
	}
}

type ReportPage struct {
	Title         string               `json:"title"`
	AlliedHosts   []ReportAlliedHost   `json:"allied_hosts"`
	OpponentHosts []ReportOpponentHost `json:"opponent_hosts"`
	SubjectUser   []ReportUser         `json:"subject_users"`
	Author        string               `json:"author"`
}

// NewReportPage is a constructor of ReportPage
func NewReportPage() ReportPage {
	page := ReportPage{}
	return page
}

type ReportResult struct {
	Severity ReportSeverity `json:"severity"`
	// Severity must be chosen from "undamaged", "unclassified", "emergency"
	//
	// urgent: Your system is damaged actually or there are strong evidence(s) of exploting system. Also incident may be on going.
	// unclassified: Not classfied and you need to check it by ownself.
	// safe: No damage by events of the alert and there is nothing to do.
}

const (
	SevHigh         ReportSeverity = "high"
	SevLow          ReportSeverity = "low"
	SevNone         ReportSeverity = "none"
	SevUnclassified ReportSeverity = "unclassified"
)

type ReportUser struct {
	UserName     string               `json:"username"` // Identity
	ServiceUsage []ReportServiceUsage `json:"service_usage"`
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
	Source   string `json:"source"`
}

type ReportDomain struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type ReportURL struct {
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type ReportServiceUsage struct {
	ServiceName string    `json:"service_name"`
	Principal   string    `json:"principal"`
	Action      string    `json:"action"`
	LastSeen    time.Time `json:"last_seen"`
}

type ReportAlliedHost struct {
	ID           string               `json:"id"`
	UserName     []string             `json:"username"`
	OS           []string             `json:"os"`
	IPAddr       []string             `json:"ipaddr"`
	MACAddr      []string             `json:"macaddr"`
	HostName     []string             `json:"hostname"`
	Country      []string             `json:"country"`
	Software     []string             `json:"software"`
	ServiceUsage []ReportServiceUsage `json:"service_usage"`
}

func (x *ReportAlliedHost) Merge(s ReportAlliedHost) {
	x.ID = s.ID
	x.UserName = append(x.UserName, s.Country...)
	x.OS = append(x.OS, s.OS...)
	x.IPAddr = append(x.IPAddr, s.IPAddr...)
	x.Country = append(x.Country, s.Country...)
	x.ServiceUsage = append(x.ServiceUsage, s.ServiceUsage...)
}

type ReportOpponentHost struct {
	ID             string          `json:"id"`
	IPAddr         []string        `json:"ipaddr"`
	Country        []string        `json:"country"`
	ASOwner        []string        `json:"as_owner"`
	RelatedMalware []ReportMalware `json:"related_malware"`
	RelatedDomains []ReportDomain  `json:"related_domains"`
	RelatedURLs    []ReportURL     `json:"related_urls"`
}

func (x *ReportOpponentHost) Merge(s ReportOpponentHost) {
	x.ID = s.ID
	x.IPAddr = append(x.IPAddr, s.IPAddr...)
	x.Country = append(x.Country, s.Country...)
	x.ASOwner = append(x.ASOwner, s.ASOwner...)
	x.RelatedMalware = append(x.RelatedMalware, s.RelatedMalware...)
	x.RelatedDomains = append(x.RelatedDomains, s.RelatedDomains...)
	x.RelatedURLs = append(x.RelatedURLs, s.RelatedURLs...)
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

// SetPage sets page data with serialization.
func (x *ReportComponent) SetPage(page ReportPage) {
	data, err := json.Marshal(&page)
	if err != nil {
		log.Println("Fail to marshal report page:", page)
	}

	x.Data = data
}

// Page returns deserialized page structure
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

	x.TimeToLive = time.Now().UTC().Add(time.Second * 864000)

	log.WithFields(log.Fields{
		"component": x,
		"tableName": tableName,
	}).Info("Put component")
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

func NewReport(reportID ReportID, alert Alert) Report {
	report := Report{
		ID:      reportID,
		Alert:   alert,
		Content: newReportContent(),
	}

	return report
}

func NewReportID() ReportID {
	return ReportID(uuid.NewV4().String())
}
