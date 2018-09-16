package lib

import (
	"io"
	"strings"
)

// CheckQPSDropFromSysbenchOutputText parses a string as sysbench output and check if there exists abnormal QPS drop.
// A record is regarded as abnormal if it's QPS is less than threshold * average_qps_of_normal_records. Returns whether
// the check passed. If not, the second return value will be a string that contains all abnormal records.
func CheckQPSDropFromSysbenchOutputText(text string, threshold float64) (bool, string) {
	reader := strings.NewReader(text)
	return CheckQPSDropFromSysbenchOutputStream(reader, threshold)
}

func CheckQPSDropFromSysbenchOutputStream(reader io.Reader, threshold float64) (bool, string) {
	rule := NewQPSDropRule(threshold)
	analyzer := NewSysbenchAnalyzer([]Rule{rule}, true)
	violations := analyzer.AnalyzeStream(reader)
	abnormalLines := ""
	for _, violation := range violations {
		abnormalLines += violation.RecordText + "\n  * " + violation.Description + "\n"
	}
	return len(violations) == 0, abnormalLines
}
