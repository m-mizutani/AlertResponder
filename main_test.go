package main_test

// This test is an integration test of AlertResponder,
// not unit test of main module.

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	// "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/guregu/dynamo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/k0kubun/pp"
	gp "github.com/m-mizutani/generalprobe"
	"github.com/m-mizutani/AlertResponder/lib"

)

var (
	Verbose        = os.Getenv("AR_VERBOSE") != ""
	TestConfigPath = os.Getenv("AR_TEST_CONFIG")
)

type testParams struct {
	StackName         string `json:"StackName"`
	Region            string `json:"Region"`
	ReportResults     string `json:"ReportResultsArn"`
}


func loadTestConfig(fpath string) testParams {
	fd, err := os.Open(fpath)
	if err != nil {
		log.Fatal("Fail to open TestConfig:", fpath, err)
	}
	defer fd.Close()

	params := testParams{}
	fdata, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal("Fail to read TestConfig:", fpath, err)
	}

	err = json.Unmarshal(fdata, &params)
	if err != nil {
		log.Fatal("Fail to unmarshal TestConfig", fpath, err)
	}

	return params
}

func getTestParams() testParams {
	var configPath string
	if TestConfigPath != "" {
		configPath = TestConfigPath
	} else {
		configPath = filepath.Join("test", "emitter", "test.json")
	}

	params := loadTestConfig(configPath)

	return params
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
	params := getTestParams()
	if Verbose {
		pp.Println("params: ", params)
	}

	alert := genAlert()
	alertData, err := json.Marshal(&alert)
	require.NoError(t, err)
	now := time.Now().UTC()

	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		gp.PublishSnsMessage("AlertNotification", alertData),
		gp.Pause(3),
		gp.GetDynamoRecord("AlertMap", func(table dynamo.Table) bool {

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

	g.Act()
}

func TestIntegration(t *testing.T) {
	params := getTestParams()
	log.WithField("params", params).Debug("Integration test")

	alert := genAlert()
	var reportID string
	var task lib.Task

	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		gp.InvokeLambda("Receptor", func(payload []byte) {
			var resp struct {
				ReportIDs []string `json:"report_ids"`
			}

			err := json.Unmarshal(payload, &resp)
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.ReportIDs))
			reportID = resp.ReportIDs[0]
		}).SnsMessage(alert),

		gp.GetKinesisStreamRecord("TaskStream", func(data []byte) bool {
			err := json.Unmarshal(data, &task)
			require.NoError(t, err)
			if string(task.ReportID) != reportID {
				return false
			}

			var page lib.ReportPage
			page.Title = "Test"
			remote := lib.ReportRemoteHost{
				IPAddr: []string{task.Attr.Value},
			}
			page.RemoteHost = append(page.RemoteHost, remote)

			cmpt := lib.NewReportComponent(lib.ReportID(reportID))
			cmpt.SetPage(page)

			reportDataTable := g.LookupID("ReportData")
			err = cmpt.Submit(reportDataTable, params.Region)
			require.NoError(t, err)
			return true
		}),

		gp.Pause(10),
		gp.GetDynamoRecord(params.ReportResults, func(table dynamo.Table) bool {
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

			assert.Equal(t, 0, len(reports[0].Content.RemoteHosts))
			assert.Equal(t, 0, len(reports[0].Content.LocalHosts))
			assert.NotEqual(t, 0, len(reports[1].Content.RemoteHosts))
			assert.Equal(t, 0, len(reports[1].Content.LocalHosts))
			return true
		}).SetTargetType(gp.ARN),
	})

	g.Act()
	return
}
