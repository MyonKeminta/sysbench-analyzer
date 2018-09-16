package lib

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

type SysbenchAnalyzer struct {
	rules              []Rule
	ignoreInvalidLines bool
}

func NewSysbenchAnalyzer(rules []Rule, ignoreInvalidLines bool) SysbenchAnalyzer {
	return SysbenchAnalyzer{
		rules:              rules,
		ignoreInvalidLines: ignoreInvalidLines,
	}
}

func (a *SysbenchAnalyzer) parseToEnd(reader io.Reader) ([]Record, []Violation) {
	var records []Record
	var violations []Violation

	scanner := bufio.NewScanner(reader)
	for {
		if !scanner.Scan() {
			break
		}
		str := scanner.Text()

		record, err := ParseRecord(str)
		if err == nil {
			records = append(records, record)
		} else {
			if !a.ignoreInvalidLines {
				violation := Violation{
					ViolatedRule: nil,
					RecordIndex:  -1,
					RecordText:   str,
					Description:  fmt.Sprintf("Invalid line. Err: %v", err),
				}
				violations = append(violations, violation)
			}
		}
	}

	return records, violations
}

func (a *SysbenchAnalyzer) AnalyzeString(str string) []Violation {
	stringReader := strings.NewReader(str)
	return a.AnalyzeStream(stringReader)
}

func (a *SysbenchAnalyzer) AnalyzeStream(reader io.Reader) []Violation {
	records, violations := a.parseToEnd(reader)
	violations = append(violations, a.AnalyzeParsedRecords(records)...)
	return violations
}

func (a *SysbenchAnalyzer) AnalyzeParsedRecords(records []Record) []Violation {
	var violations []Violation
	for _, rule := range a.rules {
		violations = append(violations, rule.Check(records)...)
	}

	sort.Slice(violations, func(i int, j int) bool {
		return violations[i].RecordIndex < violations[j].RecordIndex
	})
	return violations
}
