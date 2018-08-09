# TSBS Supplemental Guide: Cassandra

Cassandra is a general column store database. This supplemental guide explains
how the data generated for TSBS is stored, additional flags available when
using the data importer (`tsbs_load_cassandra`), and additional flags
available for the query runner (`tsbs_run_queries_cassandra`). **This
should be read *after* the main README.**

## Data format

Data generated by `tsbs_generate_data` for Cassandra is a "pseudo-CSV" format.
Each reading is a single line where the first comma-separated element with
the following elements:
* first, the table the reading belongs to (based on data type, e.g., `series_double` for doubles);
* then, the data source (e.g., `cpu` for `cpu-only`);
* then, several elements of the form `<label>=<value>` for tags;
* then, the field label;
* then, the date of the reading in YYYY-MM-DD form;
* then, the timestamp in nanoseconds;
* and finally, the reading itself.

An example from `cpu-only`:
```text
series_double,cpu,hostname=host_0,region=eu-west-1,datacenter=eu-west-1b,rack=67,os=Ubuntu16.10,arch=x86,team=NYC,service=7,service_version=0,service_environment=production,usage_guest_nice,2016-01-01,1451606400000000000,38.2431182911542820
```

When stored, the elements starting with the data source (e.g. `cpu`) through
the date of the reading are concatenated to serve as the primary key.

---

## `tsbs_load_cassandra` Additional Flags

### Database related

#### `-consistency` (type: `string`, default: `ALL`)

Consistency level for writes to the database. Options are `ALL`, `ANY`, `ONE`,
`TWO`, `THREE`, or `QUORUM`. Applies for multi-node cluster.

#### `-hosts` (type: `string`, default: `localhost:9042`)

Comma-separated list of hostname and port combinations for nodes in the cluster.

#### `-replication-factor` (type: `int`, default: `1`)

Level of replication for each write, i.e., number of nodes to store the
data on. Only applies a multi-node cluster.

#### `-write-timeout` (type: `duration`, default: `10s`)

Length of the timeout for writes.
It is expressed as a Golang time.Duration string, meaning a number followed
by a unit abbreviation (s = seconds,
m = minutes, h = hours), e.g., the default `10s` is ten seconds.


---

## `tsbs_run_queries_cassandra` Additional Flags

### Database related

#### `-aggregation-plan` (type: `string`, default: `client`)

Method for doing aggregations in queries. Due to limitations in Cassandra's
SQL-like language CQL, aggregations can be painful and slow if done on the
server itself. Therefore the default is `client` (with the other valid option
being `server`), where the client Go program handles the aggregation.

#### `-client-side-index-timeout` (type: `duration`, default: `10s`)

Length of the timeout when setting up the client side index, a data structure
used to speed up queries by storing the tagsets/primary keys in memory on the
client. It is expressed as a Golang time.Duration string, meaning a number followed by a unit abbreviation (s = seconds,
m = minutes, h = hours), e.g., the default `10s` is ten seconds.

#### `-host` (type: `string`, default: `localhost:9042`)

Hostname and port combination of at least one node in the cluster. The library
used will discover the other nodes for queries.

#### `-read-timeout` (type: `duration`, default: `10s`)

Length of the timeout for reads.
It is expressed as a Golang time.Duration string, meaning a number followed
by a unit abbreviation (s = seconds,
m = minutes, h = hours), e.g., the default `10s` is ten seconds.