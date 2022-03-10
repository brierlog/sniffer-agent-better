package exporter

import (
	"flag"
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/zr-hebo/sniffer-agent/model"
	"log"
	"os"
	"time"
)

var (
	fileSqlType  string
	fileRotation string
	fileCount    int
	filePath     string
	fileSuffix   = ".log"
)

func init() {
	flag.StringVar(
		&filePath, "file-path", "./storage/", "Save path of file. Default is ./storage/")
	flag.StringVar(
		&fileSqlType, "file-sql-type", "all", "sql type of record. Default is all")
	flag.StringVar(
		&fileRotation,
		"file-rotation", "%Y%m%d%H%M", "split time of the file. Default is minutes: {file-path}sql-audit-20060102150405.log")
	flag.IntVar(&fileCount,
		"max-file-count", 100, "maximum number of saved files. Default is 100")
}

type fileExporter struct {
}

func NewFileExporter() *fileExporter {
	return &fileExporter{}
}

func (c *fileExporter) Export(qp model.QueryPiece) (err error) {
	logier := writer()
	log.SetOutput(logier)
	log.SetFlags(0)
	switch fileSqlType {
	case "dml":
		if qp.NeedSyncDml() {
			log.Printf(*qp.String())
		}
	case "ddl":
		if qp.NeedSyncSend() {
			log.Printf(*qp.String())
		}
	default:
		log.Printf(*qp.String())
	}
	return
}

/**
文件设置
*/
func writer() *rotatelogs.RotateLogs {
	fileName := fileRotation + fileSuffix
	if _, err := os.Stat(filePath); err != nil {
		logrus.Info("Directory does not exist, creating %s", filePath)
		err := os.MkdirAll(filePath, 0755)

		if err != nil {
			panic(fmt.Sprintf("Error creating directory %s", filePath))
		}
	}
	logier, err := rotatelogs.New(
		filePath+"sql-audit-"+fileName,
		rotatelogs.WithLinkName(filePath+"sql-audit"+fileSuffix), // 生成软链，指向最新日志文件
		rotatelogs.WithRotationCount(int(fileCount)),             // 文件最大保存份数
		rotatelogs.WithRotationTime(time.Minute),                 // 日志切割时间间隔
	)

	if err != nil {
		panic(fmt.Sprintf("Error creating directory rotatelogs: %s", err.Error()))
	}
	return logier
}
