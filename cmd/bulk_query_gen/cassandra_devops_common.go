package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/440-labs/influxdb-comparisons/query"
)

// CassandraDevops produces Cassandra-specific queries for all the devops query types.
type CassandraDevops struct {
	AllInterval TimeInterval
}

// NewCassandraDevops makes an CassandraDevops object ready to generate Queries.
func newCassandraDevopsCommon(start, end time.Time) QueryGenerator {
	if !start.Before(end) {
		panic("bad time order")
	}

	return &CassandraDevops{
		AllInterval: NewTimeInterval(start, end),
	}
}

// Dispatch fulfills the QueryGenerator interface.
func (d *CassandraDevops) Dispatch(i, scaleVar int) query.Query {
	q := query.NewCassandra() // from pool
	devopsDispatchAll(d, i, q, scaleVar)
	return q
}

func (d *CassandraDevops) getHostWhere(scaleVar, nhosts int) []string {
	nn := rand.Perm(scaleVar)[:nhosts]

	tagSet := []string{}
	for _, n := range nn {
		hostname := fmt.Sprintf("host_%d", n)
		tag := fmt.Sprintf("hostname=%s", hostname)
		tagSet = append(tagSet, tag)
	}

	return tagSet
}

// TODO remove these
func (d *CassandraDevops) MaxCPUUsageHourByMinuteOneHost(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 1, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinuteTwoHosts(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 2, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinuteFourHosts(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 4, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinuteEightHosts(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 8, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinuteSixteenHosts(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 16, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinuteThirtyTwoHosts(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 32, time.Hour)
}

func (d *CassandraDevops) MaxCPUUsage12HoursByMinuteOneHost(q query.Query, scaleVar int) {
	d.maxCPUUsageHourByMinuteNHosts(q.(*query.Cassandra), scaleVar, 1, 12*time.Hour)
}

func (d *CassandraDevops) MaxCPUUsageHourByMinute(qi query.Query, scaleVar, nhosts int, timeRange time.Duration) {
	d.maxCPUUsageHourByMinuteNHosts(qi, scaleVar, nhosts, timeRange)
}

// MaxCPUUsageHourByMinuteThirtyTwoHosts populates a query.Query with a query that looks like:
// SELECT max(usage_user) from cpu where (hostname = '$HOSTNAME_1' or ... or hostname = '$HOSTNAME_N') and time >= '$HOUR_START' and time < '$HOUR_END' group by time(1m)
func (d *CassandraDevops) maxCPUUsageHourByMinuteNHosts(qi query.Query, scaleVar, nhosts int, timeRange time.Duration) {
	interval := d.AllInterval.RandWindow(timeRange)
	nn := rand.Perm(scaleVar)[:nhosts]

	tagSets := [][]string{}
	tagSet := []string{}
	for _, n := range nn {
		hostname := fmt.Sprintf("host_%d", n)
		tag := fmt.Sprintf("hostname=%s", hostname)
		tagSet = append(tagSet, tag)
	}
	tagSets = append(tagSets, tagSet)

	humanLabel := fmt.Sprintf("Cassandra max cpu, rand %4d hosts, rand %s by 1m", nhosts, timeRange)
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))

	q.AggregationType = []byte("max")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user")

	q.TimeStart = interval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Minute

	q.TagSets = tagSets
}

// CPU5Metrics selects the MAX of 5 metrics under 'cpu' per minute for nhosts hosts,
// e.g. in psuedo-SQL:
//
// SELECT minute, max(metric1), ..., max(metric5)
// FROM cpu
// WHERE (hostname = '$HOSTNAME_1' OR ... OR hostname = '$HOSTNAME_N')
// AND time >= '$HOUR_START' AND time < '$HOUR_END'
// GROUP BY minute ORDER BY minute ASC
func (d *CassandraDevops) CPU5Metrics(qi query.Query, scaleVar, nhosts int, timeRange time.Duration) {
	interval := d.AllInterval.RandWindow(timeRange)
	nn := rand.Perm(scaleVar)[:nhosts]

	tagSets := [][]string{}
	tagSet := []string{}
	for _, n := range nn {
		hostname := fmt.Sprintf("host_%d", n)
		tag := fmt.Sprintf("hostname=%s", hostname)
		tagSet = append(tagSet, tag)
	}
	tagSets = append(tagSets, tagSet)

	humanLabel := fmt.Sprintf("Cassandra 5 cpu metrics, rand %4d hosts, rand %s by 1m", nhosts, timeRange)
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))

	q.AggregationType = []byte("max")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user,usage_system,usage_idle,usage_nice,usage_guest")

	q.TimeStart = interval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Minute

	q.TagSets = tagSets
}

// GroupByOrderByLimit populates a query.Query that has a time WHERE clause, that groups by a truncated date, orders by that date, and takes a limit:
// SELECT date_trunc('minute', time) AS t, MAX(cpu) FROM cpu
// WHERE time < '$TIME'
// GROUP BY t ORDER BY t DESC
// LIMIT $LIMIT
func (d *CassandraDevops) GroupByOrderByLimit(qi query.Query, _ int) {
	interval := d.AllInterval.RandWindow(time.Hour)

	humanLabel := "Cassandra max cpu over last 5 min-intervals (rand end)"
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, d.AllInterval.StartString()))

	q.AggregationType = []byte("max")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user")

	q.TimeStart = d.AllInterval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Minute
	q.OrderBy = []byte("timestamp_ns DESC")
	q.Limit = 5
}

// MeanCPUUsageDayByHourAllHostsGroupbyHost selects the AVG of numMetrics metrics under 'cpu' per device per hour for a day,
// e.g. in psuedo-SQL:
//
// SELECT AVG(metric1), ..., AVG(metricN)
// FROM cpu
// WHERE time >= '$HOUR_START' AND time < '$HOUR_END'
// GROUP BY hour, hostname ORDER BY hour
func (d *CassandraDevops) MeanCPUUsageDayByHourAllHostsGroupbyHost(qi query.Query, numMetrics int) {
	interval := d.AllInterval.RandWindow(24 * time.Hour)
	metrics := cpuMetrics[:numMetrics]

	humanLabel := fmt.Sprintf("Cassandra mean of %d metrics, all hosts, rand 1day by 1hr", numMetrics)
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))

	q.AggregationType = []byte("avg")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte(strings.Join(metrics, ","))

	q.TimeStart = interval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Hour
}

// MaxAllCPU selects the MAX of all metrics under 'cpu' per hour for nhosts hosts,
// e.g. in psuedo-SQL:
//
// SELECT MAX(metric1), ..., MAX(metricN)
// FROM cpu WHERE (hostname = '$HOSTNAME_1' OR ... OR hostname = '$HOSTNAME_N')
// AND time >= '$HOUR_START' AND time < '$HOUR_END'
// GROUP BY hour ORDER BY hour
func (d *CassandraDevops) MaxAllCPU(qi query.Query, scaleVar, nhosts int) {
	interval := d.AllInterval.RandWindow(8 * time.Hour)
	nn := rand.Perm(scaleVar)[:nhosts]

	tagSets := [][]string{}
	tagSet := []string{}
	for _, n := range nn {
		hostname := fmt.Sprintf("host_%d", n)
		tag := fmt.Sprintf("hostname=%s", hostname)
		tagSet = append(tagSet, tag)
	}
	tagSets = append(tagSets, tagSet)

	humanLabel := fmt.Sprintf("Cassandra max cpu all fields, rand %4d hosts, rand 12hr by 1h", nhosts)
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))

	q.AggregationType = []byte("max")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user,usage_system,usage_idle,usage_nice,usage_iowait,usage_irq,usage_softirq,usage_steal,usage_guest,usage_guest_nice")

	q.TimeStart = interval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Hour

	q.TagSets = tagSets
}

func (d *CassandraDevops) LastPointPerHost(qi query.Query, _ int) {
	humanLabel := "Cassandra last row per host"
	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, d.AllInterval.StartString()))

	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user,usage_system,usage_idle,usage_nice,usage_iowait,usage_irq,usage_softirq,usage_steal,usage_guest,usage_guest_nice")

	q.TimeStart = d.AllInterval.Start
	q.TimeEnd = d.AllInterval.End

	q.ForEveryN = []byte("hostname,1")
}

// HighCPUForHosts populates a query that gets CPU metrics when the CPU has high
// usage between a time period for a number of hosts (if 0, it will search all hosts),
// e.g. in psuedo-SQL:
//
// SELECT * FROM cpu
// WHERE usage_user > 90.0
// AND time >= '$TIME_START' AND time < '$TIME_END'
// AND (hostname = '$HOST' OR hostname = '$HOST2'...)
func (d *CassandraDevops) HighCPUForHosts(qi query.Query, scaleVar, nhosts int) {
	interval := d.AllInterval.RandWindow(24 * time.Hour)
	tagSet := d.getHostWhere(scaleVar, nhosts)

	tagSets := [][]string{}
	if len(tagSet) > 0 {
		tagSets = append(tagSets, tagSet)
	}

	humanLabel := "Cassandra CPU over threshold, "
	if len(tagSet) > 0 {
		humanLabel += "one host"
	} else {
		humanLabel += "all hosts"
	}

	q := qi.(*query.Cassandra)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))

	q.AggregationType = []byte("")
	q.MeasurementName = []byte("cpu")
	q.FieldName = []byte("usage_user,usage_system,usage_idle,usage_nice,usage_iowait,usage_irq,usage_softirq,usage_steal,usage_guest,usage_guest_nice")

	q.TimeStart = interval.Start
	q.TimeEnd = interval.End
	q.GroupByDuration = time.Hour
	q.WhereClause = []byte("usage_user,>,90.0")

	q.TagSets = tagSets
}

//func (d *CassandraDevops) MeanCPUUsageDayByHourAllHostsGroupbyHost(qi query.Query, _ int) {
//	interval := d.AllInterval.RandWindow(24*time.Hour)
//
//	v := url.Values{}
//	v.Set("db", d.KeyspaceName)
//	v.Set("q", fmt.Sprintf("SELECT count(usage_user) from cpu where time >= '%s' and time < '%s' group by time(1h)", interval.StartString(), interval.EndString()))
//
//	humanLabel := "Cassandra mean cpu, all hosts, rand 1day by 1hour"
//	q := qi.(*query.Cassandra)
//	q.HumanLabel = []byte(humanLabel)
//	q.HumanDescription = []byte(fmt.Sprintf("%s: %s", humanLabel, interval.StartString()))
//	q.Method = []byte("GET")
//	q.Path = []byte(fmt.Sprintf("/query?%s", v.Encode()))
//	q.Body = nil
//}
