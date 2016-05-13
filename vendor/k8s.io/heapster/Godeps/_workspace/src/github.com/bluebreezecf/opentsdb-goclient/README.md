##opentsdb-goclient

###Backgroud
OpenTSDB is a distributed, scalable Time Series Database (TSDB) written on top of HBase.
OpenTSDB was written to address a common need: store, index and serve metrics collected 
from computer systems (network gear, operating systems, applications) at a large scale, 
and make this data easily accessible and graphable.

I am about to use OpenTSDB, but currently there is no useable go-sdk for OpenTSDB. So I 
develop the opentsdb-goclient for convenience according to the [OpenTSDB Rest API Doc] (http://opentsdb.net/docs/build/html/api_http/index.html#api-endpoints)

###How to use sample
If you want to see how does the sample (sample.go) run, you can execute the following commands:
```shell
cd $GOPATH
mkdir -p $GOPATH/src/github.com/bluebreezecf
cd $GOPATH/src/github.com/bluebreezecf
git clone https://github.com/bluebreezecf/opentsdb-goclient.git

vi sample.go //Use the real host and port of an existing OpenTSDB in Line 33: OpentsdbHost: "127.0.0.1:4242"
go run sample.go

```

###Current supporting rest apis
```shell
GET             /api/aggregators
GET,POST,DELETE /api/annotation
POST,DELETE     /api/annotation/bulk
GET 			/api/config
GET 			/api/dropcaches
POST 			/api/put
GET				/api/query
GET 			/api/serializers
GET 			/api/stats
GET				/api/suggest
POST			/api/uid/assign
GET,POST,DELETE /api/uid/tsmeta
GET,POST,DELETE /api/uid/uidmeta
GET 			/api/version
```