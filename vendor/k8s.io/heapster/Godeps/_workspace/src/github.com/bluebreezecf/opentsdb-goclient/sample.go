// Copyright 2015 opentsdb-goclient authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// Package main shows the sample of how to use github.com/bluebreezecf/opentsdbclient/client
// to communicate with the OpenTSDB with the pre-define rest apis.
// (http://opentsdb.net/docs/build/html/api_http/index.html#api-endpoints)
//
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bluebreezecf/opentsdb-goclient/client"
	"github.com/bluebreezecf/opentsdb-goclient/config"
)

func main() {
	opentsdbCfg := config.OpenTSDBConfig{
		OpentsdbHost: "127.0.0.1:4242",
	}
	tsdbClient, err := client.NewClient(opentsdbCfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//0. Ping
	if err = tsdbClient.Ping(); err != nil {
		fmt.Println(err.Error())
		return
	}
	PutDataPointNum := 4

	//1. POST /api/put
	fmt.Println("Begin to test POST /api/put.")
	cpuDatas := make([]client.DataPoint, 0)
	st1 := time.Now().Unix()
	time.Sleep(2 * time.Second)
	tags := make(map[string]string)
	tags["host"] = "bluebreezecf-host"
	i := 0
	for {
		time.Sleep(500 * time.Millisecond)
		data := client.DataPoint{
			Metric:    "cpu",
			Timestamp: time.Now().Unix(),
			Value:     rand.Float64(),
		}
		tags := make(map[string]string)
		tags["host"] = "bluebreezecf-host"
		data.Tags = tags
		cpuDatas = append(cpuDatas, data)
		fmt.Printf("  %d.Prepare datapoint %s\n", i, data.String())
		if i < PutDataPointNum {
			i++
		} else {
			break
		}
	}

	if resp, err := tsdbClient.Put(cpuDatas, "details"); err != nil {
		fmt.Printf("  Error occurs when putting datapoints: %v", err)
	} else {
		fmt.Printf("  %s", resp.String())
	}
	fmt.Println("Finish testing POST /api/put.")

	//2. POST /api/query
	fmt.Println("Begin to test POST /api/query.")
	st2 := time.Now().Unix()
	queryParam := client.QueryParam{
		Start: st1,
		End:   st2,
	}
	subqueries := make([]client.SubQuery, 0)
	subQuery := client.SubQuery{
		Aggregator: "sum",
		Metric:     "cpu",
		Tags:       tags,
	}
	subqueries = append(subqueries, subQuery)
	queryParam.Queries = subqueries
	if queryResp, err := tsdbClient.Query(queryParam); err != nil {
		fmt.Printf("Error occurs when querying: %v", err)
	} else {
		fmt.Printf("%s", queryResp.String())
	}
	fmt.Println("Finish testing POST /api/query.")

	//3. GET /api/aggregators
	fmt.Println("Begin to test GET /api/aggregators.")
	aggreResp, err := tsdbClient.Aggregators()
	if err != nil {
		fmt.Printf("Error occurs when acquiring aggregators: %v", err)
	} else {
		fmt.Printf("%s", aggreResp.String())
	}
	fmt.Println("Finish testing GET /api/aggregators.")

	//4. GET /api/config
	fmt.Println("Begin to test GET /api/config.")
	configResp, err := tsdbClient.Config()
	if err != nil {
		fmt.Printf("Error occurs when acquiring config info: %v", err)
	} else {
		fmt.Printf("%s", configResp.String())
	}
	fmt.Println("Finish testing GET /api/config.")

	//5. Get /api/serializers
	fmt.Println("Begin to test GET /api/serializers.")
	serilResp, err := tsdbClient.Serializers()
	if err != nil {
		fmt.Printf("Error occurs when acquiring serializers info: %v", err)
	} else {
		fmt.Printf("%s", serilResp.String())
	}
	fmt.Println("Finish testing GET /api/serializers.")

	//6. Get /api/stats
	fmt.Println("Begin to test GET /api/stats.")
	statsResp, err := tsdbClient.Stats()
	if err != nil {
		fmt.Printf("Error occurs when acquiring stats info: %v", err)
	} else {
		fmt.Printf("%s", statsResp.String())
	}
	fmt.Println("Finish testing GET /api/stats.")

	//7. Get /api/suggest
	fmt.Println("Begin to test GET /api/suggest.")
	typeValues := []string{client.TypeMetrics, client.TypeTagk, client.TypeTagv}
	for _, typeItem := range typeValues {
		sugParam := client.SuggestParam{
			Type: typeItem,
		}
		fmt.Printf("  Send suggest param: %s", sugParam.String())
		sugResp, err := tsdbClient.Suggest(sugParam)
		if err != nil {
			fmt.Printf("  Error occurs when acquiring suggest info: %v\n", err)
		} else {
			fmt.Printf("  Recevie response: %s\n", sugResp.String())
		}
	}
	fmt.Println("Finish testing GET /api/suggest.")

	//8. Get /api/version
	fmt.Println("Begin to test GET /api/version.")
	versionResp, err := tsdbClient.Version()
	if err != nil {
		fmt.Printf("Error occurs when acquiring version info: %v", err)
	} else {
		fmt.Printf("%s", versionResp.String())
	}
	fmt.Println("Finish testing GET /api/version.")

	//9. Get /api/dropcaches
	fmt.Println("Begin to test GET /api/dropcaches.")
	dropResp, err := tsdbClient.Dropcaches()
	if err != nil {
		fmt.Printf("Error occurs when acquiring dropcaches info: %v", err)
	} else {
		fmt.Printf("%s", dropResp.String())
	}
	fmt.Println("Finish testing GET /api/dropcaches.")

	//10. POST /api/annotation
	fmt.Println("Begin to test POST /api/annotation.")
	custom := make(map[string]string, 0)
	custom["owner"] = "bluebreezecf"
	custom["host"] = "bluebreezecf-host"
	addedST := time.Now().Unix()
	addedTsuid := "000001000001000002"
	anno := client.Annotation{
		StartTime:   addedST,
		Tsuid:       addedTsuid,
		Description: "bluebreezecf test annotation",
		Notes:       "These would be details about the event, the description is just a summary",
		Custom:      custom,
	}
	if queryAnnoResp, err := tsdbClient.UpdateAnnotation(anno); err != nil {
		fmt.Printf("Error occurs when posting annotation info: %v", err)
	} else {
		fmt.Printf("%s", queryAnnoResp.String())
	}
	fmt.Println("Finish testing POST /api/annotation.")

	//11. GET /api/annotation
	fmt.Println("Begin to test GET /api/annotation.")
	queryAnnoMap := make(map[string]interface{}, 0)
	queryAnnoMap[client.AnQueryStartTime] = addedST
	queryAnnoMap[client.AnQueryTSUid] = addedTsuid
	if queryAnnoResp, err := tsdbClient.QueryAnnotation(queryAnnoMap); err != nil {
		fmt.Printf("Error occurs when acquiring annotation info: %v", err)
	} else {
		fmt.Printf("%s", queryAnnoResp.String())
	}
	fmt.Println("Finish testing GET /api/annotation.")

	//12. GET /api/annotation
	fmt.Println("Begin to test DELETE /api/annotation.")
	if queryAnnoResp, err := tsdbClient.DeleteAnnotation(anno); err != nil {
		fmt.Printf("Error occurs when deleting annotation info: %v", err)
	} else {
		fmt.Printf("%s", queryAnnoResp.String())
	}
	fmt.Println("Finish testing DELETE /api/annotation.")

	//13. POST /api/annotation/bulk
	fmt.Println("Begin to test POST /api/annotation/bulk.")
	anns := make([]client.Annotation, 0)
	bulkAnnNum := 4
	i = 0
	bulkAddBeginST := time.Now().Unix()
	addedTsuids := make([]string, bulkAnnNum)
	for {
		if i < bulkAnnNum-1 {
			addedST := time.Now().Unix()
			addedTsuid := fmt.Sprintf("%s%d", "00000100000100000", i)
			addedTsuids = append(addedTsuids, addedTsuid)
			anno := client.Annotation{
				StartTime:   addedST,
				Tsuid:       addedTsuid,
				Description: "bluebreezecf test annotation",
				Notes:       "These would be details about the event, the description is just a summary",
			}
			anns = append(anns, anno)
			i++
		} else {
			break
		}
	}
	if bulkAnnoResp, err := tsdbClient.BulkUpdateAnnotations(anns); err != nil {
		fmt.Printf("Error occurs when posting bulk annotation info: %v", err)
	} else {
		fmt.Printf("%s", bulkAnnoResp.String())
	}
	fmt.Println("Finish testing POST /api/annotation/bulk.")

	//14. DELETE /api/annotation/bulk
	fmt.Println("Begin to test DELETE /api/annotation/bulk.")
	bulkAnnoDelete := client.BulkAnnoDeleteInfo{
		StartTime: bulkAddBeginST,
		Tsuids:    addedTsuids,
		Global:    false,
	}
	if bulkAnnoResp, err := tsdbClient.BulkDeleteAnnotations(bulkAnnoDelete); err != nil {
		fmt.Printf("Error occurs when deleting bulk annotation info: %v", err)
	} else {
		fmt.Printf("%s", bulkAnnoResp.String())
	}
	fmt.Println("Finish testing DELETE /api/annotation/bulk.")

	//15. GET /api/uid/uidmeta
	fmt.Println("Begin to test GET /api/uid/uidmeta.")
	metaQueryParam := make(map[string]string, 0)
	metaQueryParam["type"] = client.TypeMetrics
	metaQueryParam["uid"] = "00003A"
	if resp, err := tsdbClient.QueryUIDMetaData(metaQueryParam); err != nil {
		fmt.Printf("Error occurs when querying uidmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}

	fmt.Println("Finish testing GET /api/uid/uidmeta.")

	//16. POST /api/uid/uidmeta
	fmt.Println("Begin to test POST /api/uid/uidmeta.")
	uidMetaData := client.UIDMetaData{
		Uid:         "00002A",
		Type:        "metric",
		DisplayName: "System CPU Time",
	}
	if resp, err := tsdbClient.UpdateUIDMetaData(uidMetaData); err != nil {
		fmt.Printf("Error occurs when posting uidmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing POST /api/uid/uidmeta.")

	//17. DELETE /api/uid/uidmeta
	fmt.Println("Begin to test DELETE /api/uid/uidmeta.")
	uidMetaData = client.UIDMetaData{
		Uid:  "00003A",
		Type: "metric",
	}
	if resp, err := tsdbClient.DeleteUIDMetaData(uidMetaData); err != nil {
		fmt.Printf("Error occurs when deleting uidmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing DELETE /api/uid/uidmeta.")

	//18. POST /api/uid/assign
	fmt.Println("Begin to test POST /api/uid/assign.")
	metrics := []string{"sys.cpu.0", "sys.cpu.1", "illegal!character"}
	tagk := []string{"host"}
	tagv := []string{"web01", "web02", "web03"}
	assignParam := client.UIDAssignParam{
		Metric: metrics,
		Tagk:   tagk,
		Tagv:   tagv,
	}
	if resp, err := tsdbClient.AssignUID(assignParam); err != nil {
		fmt.Printf("Error occurs when assgining uid info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing POST /api/uid/assign.")

	//19. GET /api/uid/tsmeta
	fmt.Println("Begin to test GET /api/uid/tsmeta.")
	if resp, err := tsdbClient.QueryTSMetaData("000001000001000001"); err != nil {
		fmt.Printf("Error occurs when querying tsmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing GET /api/uid/tsmeta.")

	//20. POST /api/uid/tsmeta
	fmt.Println("Begin to test POST /api/uid/tsmeta.")
	custom = make(map[string]string, 0)
	custom["owner"] = "bluebreezecf"
	custom["department"] = "paas dep"
	tsMetaData := client.TSMetaData{
		Tsuid:       "000001000001000001",
		DisplayName: "System CPU Time for Webserver 01",
		Custom:      custom,
	}
	if resp, err := tsdbClient.UpdateTSMetaData(tsMetaData); err != nil {
		fmt.Printf("Error occurs when posting tsmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing POST /api/uid/tsmeta.")

	//21. DELETE /api/uid/tsmeta
	fmt.Println("Begin to test DELETE /api/uid/tsmeta.")
	tsMetaData = client.TSMetaData{
		Tsuid: "000001000001000001",
	}
	if resp, err := tsdbClient.DeleteTSMetaData(tsMetaData); err != nil {
		fmt.Printf("Error occurs when deleting tsmetadata info: %v", err)
	} else {
		fmt.Printf("%s", resp.String())
	}
	fmt.Println("Finish testing DELETE /api/uid/tsmeta.")
}
