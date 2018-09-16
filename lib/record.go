package lib

import (
	"fmt"
)

type Record struct {
	Second          int32
	Threads         int32
	TPS             float64
	QPS             float64
	ReadQPS         float64
	WriteQPS        float64
	OtherQPS        float64
	Latency         float64
	ErrorPerSec     float64
	ReconnectPerSec float64
}

const (
	recordParseFormat = "[ %ds ] thds: %d tps: %f qps: %f (r/w/o: %f/%f/%f) lat (ms,95%%): %f err/s: %f reconn/s: %f"
	recordPrintFormat = "[ %ds ] thds: %d tps: %.02f qps: %.02f (r/w/o: %.02f/%.02f/%.02f) lat (ms,95%%): %.02f err/s: %.02f reconn/s: %.02f"
)

func ParseRecord(str string) (Record, error) {
	record := Record{}

	_, err := fmt.Sscanf(
		str,
		recordParseFormat,
		&record.Second,
		&record.Threads,
		&record.TPS,
		&record.QPS,
		&record.ReadQPS,
		&record.WriteQPS,
		&record.OtherQPS,
		&record.Latency,
		&record.ErrorPerSec,
		&record.ReconnectPerSec)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}

func FormatRecord(record Record) string {
	return fmt.Sprintf(
		recordPrintFormat,
		record.Second,
		record.Threads,
		record.TPS,
		record.QPS,
		record.ReadQPS,
		record.WriteQPS,
		record.OtherQPS,
		record.Latency,
		record.ErrorPerSec,
		record.ReconnectPerSec)
}
