package lib

import (
	. "github.com/pingcap/check"
)

type testRulesSuite struct{}

var _ = Suite(&testRulesSuite{})

func (s *testRulesSuite) TestQPSDropRule(c *C) {
	rule := NewQPSDropRule(0.7)
	analyzer := NewSysbenchAnalyzer([]Rule{rule}, true)

	normalRecords1 := `[ 2s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 4s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 6s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 8s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 12s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 14s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
`
	res := analyzer.AnalyzeString(normalRecords1)
	c.Assert(len(res), Equals, 0)

	normalRecords2 := `[ 2s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 4s ] thds: 256 tps: 2085.94 qps: 28001.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 6s ] thds: 256 tps: 2085.94 qps: 28001.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 8s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 12s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 14s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
`
	res = analyzer.AnalyzeString(normalRecords2)
	c.Assert(len(res), Equals, 0)

	abnormalRecords1 := `[ 2s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 4s ] thds: 256 tps: 2085.94 qps: 27999.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 6s ] thds: 256 tps: 2085.94 qps: 27999.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 8s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 12s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
[ 14s ] thds: 256 tps: 2085.94 qps: 40000.00 (r/w/o: 30721.33/8665.87/4301.72) lat (ms,95%): 223.34 err/s: 0.00 reconn/s: 0.00
`
	res = analyzer.AnalyzeString(abnormalRecords1)
	c.Assert(len(res), Equals, 2)
	c.Assert(res[0].RecordIndex, Equals, 1)
	c.Assert(res[1].RecordIndex, Equals, 2)
	c.Assert(res[0].ViolatedRule, Equals, rule)
	c.Assert(res[1].ViolatedRule, Equals, rule)
}
