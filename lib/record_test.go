package lib

import (
	. "github.com/pingcap/check"
	"testing"
)

func TestT(t *testing.T) {
	TestingT(t)
}

type testRecordSuite struct{}

var _ = Suite(&testRecordSuite{})

func (s *testRecordSuite) TestParsing(c *C) {
	records := []string{
		"[ 8s ] thds: 256 tps: 2090.02 qps: 41970.30 (r/w/o: 29361.71/8421.56/4187.03) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00",
		"[ 10s ] thds: 256 tps: 0.00 qps: 0.00 (r/w/o: 0.00/0.00/0.00) lat (ms,95%): 0.00 err/s: 0.00 reconn/s: 0.00",
	}

	for _, record := range records {
		parsed, err := ParseRecord(record)
		c.Assert(err, IsNil)
		c.Assert(FormatRecord(parsed), Equals, record)
	}

	invalidRecords := []string{
		"[ 8s ] thds: 256 tps: 2090.02 qps: 41970.30 (r/w/o: 29361.71/8421.56/4187.03) lat (ms,95%): 223.34 err/s: 0.00 reconn/s:",
		"[ 8s ] thds: 256 tps: 2090.02 qps: 41970.30 (r/w/o: 29361.71/8421.56/4187.03) lat (ms,95%): 223.34 err/s: 0.00 reconn/ss: 0.00",
	}

	for _, record := range invalidRecords {
		_, err := ParseRecord(record)
		c.Assert(err, NotNil)
	}
}
