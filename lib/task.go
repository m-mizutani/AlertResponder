package lib

type Task struct {
	Attr     Attribute `json:"attribute"`
	ReportID ReportID  `json:"report_id"`
	Alert    Alert     `json:"alert"`
}
