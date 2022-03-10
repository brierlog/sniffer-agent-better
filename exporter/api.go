package exporter

import (
	"bytes"
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/zr-hebo/sniffer-agent/model"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	apiReportUrl string
	apiType      string
	httpClient   = &http.Client{Timeout: 5 * time.Second}
)

type mysqlSqlResult struct {
	Cip   string `json:"cip"`
	Cport int    `json:"cport"`
	User  string `json:"user"`
	Db    string `json:"db"`
	Sql   string `json:"sql"`
	Bt    int64  `json:"bt"`
	Cms   int64  `json:"cms"`
}
type sqlReportDate struct {
	ClientIp    string `json:"client_ip"`
	ClientPort  int    `json:"client_port"`
	User        string `json:"user"`
	Db          string `json:"db"`
	Sql         string `json:"sql"`
	SqlType     string `json:"sql_type"`
	StartTime   int    `json:"start_time"`
	EndTime     int    `json:"end_time"`
	ProjectId   string `json:"project_id"`
	ClusterId   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
}

func init() {
	flag.StringVar(
		&apiType, "api-sql-type", "dml|ddl", "sql type of record. Default is dml、ddl")
	flag.StringVar(
		&apiReportUrl, "api-report-url", "http://127.0.0.1/report", "data reporting api address. Default is http://127.0.0.1/report")
}

type apiExporter struct {
}

func NewApiExporter() *apiExporter {

	return &apiExporter{}
}

func (c *apiExporter) Export(qp model.QueryPiece) (err error) {
	// 超时时间：5秒
	contentType := "application/json"
	rspData, err := getReportData(*qp.String())
	if err != nil {
		log.Errorf("export with api failed -- %s", err.Error())
		return
	}
	switch apiType {
	case "dml":
		if qp.NeedSyncDml() {
			rspData.SqlType = "DML"
			_, err := post(apiReportUrl, rspData, contentType)
			if err != nil {
				log.Errorf("export with api failed -- %s", err.Error())
			}
			log.Debugln(*qp.String())
		}
	case "ddl":
		if qp.NeedSyncSend() {
			rspData.SqlType = "DDL"
			_, err := post(apiReportUrl, rspData, contentType)
			if err != nil {
				log.Errorf("export with api failed -- %s", err.Error())
			}
			log.Debugln(*qp.String())
		}
	case "dml|ddl":
		if qp.NeedSyncDml() {
			rspData.SqlType = "DML"
			_, err := post(apiReportUrl, rspData, contentType)
			if err != nil {
				log.Errorf("export with api failed -- %s", err.Error())
			}
			log.Debugln(*qp.String())
		}
		if qp.NeedSyncSend() {
			rspData.SqlType = "DDL"
			_, err := post(apiReportUrl, rspData, contentType)
			if err != nil {
				log.Errorf("export with api failed -- %s", err.Error())
			}
			log.Debugln(*qp.String())
		}
	default:
		_, err := post(apiReportUrl, rspData, contentType)
		if err != nil {
			log.Errorf("export with api failed -- %s", err.Error())
		}
		log.Debugln(*qp.String())
	}
	return
}

func getReportData(mqpStr string) (sqlReportDate, error) {
	sqlRes := mysqlSqlResult{}
	err := json.Unmarshal([]byte(mqpStr), &sqlRes)
	if err != nil {
		return sqlReportDate{}, err
	}
	startTime := int(sqlRes.Bt / 1000)
	endTime := startTime
	if cms := int(sqlRes.Cms / 1000); cms > 0 {
		endTime = startTime + cms
	}
	sqlRep := sqlReportDate{
		ClientIp:    sqlRes.Cip,
		ClientPort:  sqlRes.Cport,
		User:        sqlRes.User,
		Db:          sqlRes.Db,
		Sql:         sqlRes.Sql,
		StartTime:   startTime,
		EndTime:     endTime,
		ProjectId:   os.Getenv("__MIGU_PROJECT_ID__"),
		ClusterId:   os.Getenv("__MIGU_CLUSTER_ID__"),
		ClusterName: os.Getenv("__MIGU_CLUSTER_NAME__"),
	}
	return sqlRep, nil
}

// Post 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
func post(url string, data sqlReportDate, contentType string) (string, error) {
	jsonStr, _ := json.Marshal(data)
	resp, err := httpClient.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return string(result), nil
}
