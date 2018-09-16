package lib

type Violation struct {
	ViolatedRule Rule
	RecordIndex  int
	RecordText   string
	Description  string
}
