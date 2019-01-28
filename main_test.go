package main_test

// This test is an integration test of AlertResponder,
// not unit test of main module.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	// "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/guregu/dynamo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/m-mizutani/generalprobe"
)

var (
	Verbose        = os.Getenv("AR_VERBOSE") != ""
	TestConfigPath = os.Getenv("AR_TEST_CONFIG")
)

type testerParams struct {
	Region        string `json:"Region"`
	AccountId     string `json:"AccountId"`
	InspectorName string `json:"Inspector"`
	ReporterName  string `json:"Reporter"`
}

type mainParams struct {
	StackName string `json:"StackName"`
	Region    string `json:"Region"`
}

func loadParameters(fpath string, params interface{}) {
	fd, err := os.Open(fpath)
	if err != nil {
		log.Fatal("Fail to open a parameter file:", fpath, err)
	}
	defer fd.Close()

	fdata, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal("Fail to read a parameter file:", fpath, err)
	}

	err = json.Unmarshal(fdata, &params)
	if err != nil {
		log.Fatal("Fail to unmarshal a parameter file", fpath, err)
	}
}

func getTesterParameters() testerParams {
	testerPath := filepath.Join("tester", "params.json")
	if TestConfigPath != "" {
		testerPath = TestConfigPath
	}

	var params testerParams
	loadParameters(testerPath, &params)
	return params
}

func getMainParameters() mainParams {
	var mainParams mainParams
	loadParameters(os.Getenv("AR_CONFIG"), &mainParams)
	return mainParams
}

type reportResult struct {
	ReportID  lib.ReportID `dynamo:"report_id"`
	Timestamp float64      `dynamo:"timestamp"`
	Report    []byte       `dynamo:"report"`
}

func genAlert() lib.Alert {
	alertKey := uuid.NewV4().String()
	if Verbose {
		pp.Println("AlertKey = ", alertKey)
	}

	alert := lib.Alert{
		Name: "Test Detection",
		Key:  alertKey,
		Rule: "test",
		Attrs: []lib.Attribute{
			lib.Attribute{
				Type:    "ipaddr",
				Value:   "195.22.26.248",
				Key:     "source address",
				Context: []string{"remote"},
			},
		},
	}

	return alert
}

func TestInvokeBySns(t *testing.T) {
	params := getMainParameters()
	if Verbose {
		pp.Println("params: ", params)
	}

	alert := genAlert()
	alertData, err := json.Marshal(&alert)
	require.NoError(t, err)
	now := time.Now().UTC()

	g := generalprobe.New(params.Region, params.StackName)
	g.AddScenes([]generalprobe.Scene{
		g.PublishSnsMessage(g.LogicalID("AlertNotification"), alertData),
		g.Pause(3),
		g.GetDynamoRecord(g.LogicalID("AlertMap"), func(table dynamo.Table) bool {

			type AlertRecord struct {
				AlertID   string       `dynamo:"alert_id"`
				AlertKey  string       `dynamo:"alert_key"`
				Rule      string       `dynamo:"rule"`
				ReportID  lib.ReportID `dynamo:"report_id"`
				AlertData []byte       `dynamo:"alert_data"`
				Timestamp time.Time    `dynamo:"timestamp"`
				TTL       time.Time    `dynamo:"ttl"`
			}

			var results []AlertRecord

			err := table.Scan().Filter("'timestamp' >= ?", now).All(&results)
			if err != nil {
				return false
			}

			if len(results) == 0 {
				return false
			}

			assert.Equal(t, 1, len(results))
			return true
		}),
	})

	g.Run()
}

func TestIntegration(t *testing.T) {
	params := getMainParameters()
	tester := getTesterParameters()

	log.WithField("params", params).Debug("Integration test")

	alert := genAlert()
	var reportID string
	// var task lib.Task

	inspectorArn := fmt.Sprintf("arn:aws:lambda:%s:%s:%s", tester.Region,
		tester.AccountId, tester.InspectorName)
	reporterArn := fmt.Sprintf("arn:aws:lambda:%s:%s:%s", tester.Region,
		tester.AccountId, tester.ReporterName)

	hasReportID := func(logs []string) bool {
		pp.Println(logs)
		for _, line := range logs {
			if strings.Index(line, reportID) >= 0 {
				return true
			}
		}

		return false
	}

	g := generalprobe.New(params.Region, params.StackName)
	g.AddScenes([]generalprobe.Scene{
		g.InvokeLambda(g.LogicalID("Receptor"), func(payload []byte) {
			var resp struct {
				ReportIDs []string `json:"report_ids"`
			}

			err := json.Unmarshal(payload, &resp)
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.ReportIDs))
			reportID = resp.ReportIDs[0]
			pp.Println(resp)
		}).SnsMessage(alert),

		g.GetLambdaLogs(g.LogicalID("Dispatcher"), "", hasReportID),
		g.GetLambdaLogs(g.Arn(inspectorArn), "", hasReportID),
		g.Pause(10),
		g.GetLambdaLogs(g.LogicalID("Compiler"), "", hasReportID),
		g.GetLambdaLogs(g.LogicalID("Publisher"), "", hasReportID),
		g.GetLambdaLogs(g.Arn(reporterArn), "", hasReportID),

		/*
			g.GetKinesisStreamRecord(g.LogicalID("TaskStream"), func(data []byte) bool {
				err := json.Unmarshal(data, &task)
				require.NoError(t, err)
				if string(task.ReportID) != reportID {
					return false
				}

				var page lib.ReportPage
				page.Title = "Test"
				remote := lib.ReportOpponentHost{
					IPAddr: []string{task.Attr.Value},
				}
				page.OpponentHosts = append(page.OpponentHosts, remote)

				cmpt := lib.NewReportComponent(lib.ReportID(reportID))
				cmpt.SetPage(page)

				reportDataTable := g.LookupID("ReportData")
				err = cmpt.Submit(reportDataTable, params.Region)
				require.NoError(t, err)
				return true
			}),
		*/

		/*
			g.GetDynamoRecord(g.Arn(params.ReportResults), func(table dynamo.Table) bool {
				// Check eventual result(s)
				var results []reportResult
				err := table.Get("report_id", reportID).All(&results)
				require.NoError(t, err)
				if 2 != len(results) {
					return false
				}

				reports := []lib.Report{}

				sort.Slice(results, func(i, j int) bool { return results[i].Timestamp < results[j].Timestamp })

				for _, result := range results {
					var report lib.Report
					err := json.Unmarshal(result.Report, &report)
					assert.NoError(t, err)
					reports = append(reports, report)
				}

				assert.Equal(t, 0, len(reports[0].Content.OpponentHosts))
				assert.Equal(t, 0, len(reports[0].Content.AlliedHosts))
				assert.NotEqual(t, 0, len(reports[1].Content.OpponentHosts))
				assert.Equal(t, 0, len(reports[1].Content.AlliedHosts))
				return true
			}),
		*/
	})

	g.Run()
	return
}
