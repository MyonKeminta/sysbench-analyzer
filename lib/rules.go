package lib

import (
	"fmt"
	"sort"
)

type Rule interface {
	Check(records []Record) []Violation
	GetDescription() string
}

type QPSDropRule struct {
	threshold float64
}

// NewQPSDropRule creates a rule that checks if there is a set of records with their QPS all less than threshold * average_of_others.
// No more than half of records will be put into the set.
func NewQPSDropRule(threshold float64) Rule {
	return &QPSDropRule{
		threshold: threshold,
	}
}

func (r *QPSDropRule) Check(records []Record) []Violation {
	// Do not check it if if records are too few
	if len(records) < 5 {
		return []Violation{}
	}

	// `increasingIndex` saves a list of index pointing to elements in `records`, and it is sorted by index of it's
	// corresponding record's QPS in increasing order.
	increasingIndex := make([]int, len(records))
	for i := 0; i < len(increasingIndex); i++ {
		increasingIndex[i] = i
	}
	sort.Slice(increasingIndex, func(i int, j int) bool {
		return records[increasingIndex[i]].QPS < records[increasingIndex[j]].QPS
	})

	// qpsSum is the average QPS among records[increasingIndex[i..]]
	qpsSum := 0.0
	for _, record := range records {
		qpsSum += record.QPS
	}

	maxViolatedIndex := -1
	maxViolationAverage := 0.0
	for i := 0; i <= len(increasingIndex)/2; i++ {
		qpsSum -= records[increasingIndex[i]].QPS
		avg := qpsSum / float64(len(records)-i-1)
		if avg > 0 && records[increasingIndex[i]].QPS < r.threshold*avg {
			maxViolatedIndex = i
			maxViolationAverage = avg
		}
	}

	var violations []Violation
	for i := 0; i <= maxViolatedIndex; i++ {
		record := records[increasingIndex[i]]
		violation := Violation{
			ViolatedRule: r,
			RecordIndex:  increasingIndex[i],
			RecordText:   FormatRecord(record),
			Description: fmt.Sprintf(
				"The QPS (%.02f) is less than %.1f%% of the average QPS of normal records (%.02f)",
				record.QPS,
				r.threshold*100,
				maxViolationAverage),
		}
		violations = append(violations, violation)
	}
	return violations
}

func (r *QPSDropRule) GetDescription() string {
	return "QPSDropRule: Check if there's a record with its QPS less than average of others."
}
