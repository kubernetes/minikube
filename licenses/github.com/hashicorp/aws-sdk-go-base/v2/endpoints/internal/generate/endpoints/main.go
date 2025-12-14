// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/aws-sdk-go-base/v2/internal/generate/common"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/slices"
)

type PartitionDatum struct {
	ID          string
	Name        string
	DNSSuffix   string
	RegionRegex string
	Regions     []RegionDatum
	Services    []ServiceDatum
}

type RegionDatum struct {
	ID          string
	Description string
}

type ServiceDatum struct {
	ID string
}

type TemplateData struct {
	Partitions []PartitionDatum
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go <aws-sdk-go-v2-endpoints-json-url>\n\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	inputURL := args[0]
	filename := `endpoints_gen.go`
	target := map[string]any{}

	g := common.NewGenerator()
	g.Infof("Generating endpoints/%s", filename)

	if err := readHTTPJSON(inputURL, &target); err != nil {
		g.Fatalf("error reading JSON from %s: %s", inputURL, err)
	}

	td := TemplateData{}
	templateFuncMap := template.FuncMap{
		// IDToTitle splits a '-' or '.' separated string and returns a string with each part title cased.
		"IDToTitle": func(s string) (string, error) {
			parts := strings.Split(s, "-")
			if len(parts) == 1 {
				parts = strings.Split(s, ".")
			}
			return strings.Join(slices.ApplyToAll(parts, func(s string) string {
				return common.FirstUpper(s)
			}), ""), nil
		},
	}

	if version, ok := target["version"].(float64); ok {
		if version != 3.0 {
			g.Fatalf("unsupported endpoints document version: %d", int(version))
		}
	} else {
		g.Fatalf("can't parse endpoints document version")
	}

	/*
		See https://github.com/aws/aws-sdk-go/blob/main/aws/endpoints/v3model.go.
		e.g.
		{
		  "partitions": [{
		    "partition": "aws",
		    "partitionName": "AWS Standard",
		    "regions" : {
		      "af-south-1" : {
		        "description" : "Africa (Cape Town)"
		      },
			  ...
		    },
		    "services" : {
		      "access-analyzer" : {
		        "endpoints" : {
		          "af-south-1" : { },
				  ...
		        },
		      },
		      ...
		    },
			...
		   }, ...]
		}
	*/
	if partitions, ok := target["partitions"].([]any); ok {
		for _, partition := range partitions {
			if partition, ok := partition.(map[string]any); ok {
				partitionDatum := PartitionDatum{}

				if id, ok := partition["partition"].(string); ok {
					partitionDatum.ID = id
				}
				if name, ok := partition["partitionName"].(string); ok {
					partitionDatum.Name = name
				}
				if dnsSuffix, ok := partition["dnsSuffix"].(string); ok {
					partitionDatum.DNSSuffix = dnsSuffix
				}
				if regionRegex, ok := partition["regionRegex"].(string); ok {
					partitionDatum.RegionRegex = regionRegex
				}
				if regions, ok := partition["regions"].(map[string]any); ok {
					for id, region := range regions {
						regionDatum := RegionDatum{
							ID: id,
						}

						if region, ok := region.(map[string]any); ok {
							if description, ok := region["description"].(string); ok {
								regionDatum.Description = description
							}
						}

						partitionDatum.Regions = append(partitionDatum.Regions, regionDatum)
					}
				}
				if services, ok := partition["services"].(map[string]any); ok {
					for id := range services {
						serviceDatum := ServiceDatum{
							ID: id,
						}

						partitionDatum.Services = append(partitionDatum.Services, serviceDatum)
					}
				}

				td.Partitions = append(td.Partitions, partitionDatum)
			}
		}
	}

	sort.SliceStable(td.Partitions, func(i, j int) bool {
		return td.Partitions[i].ID < td.Partitions[j].ID
	})

	for i := 0; i < len(td.Partitions); i++ {
		sort.SliceStable(td.Partitions[i].Regions, func(j, k int) bool {
			return td.Partitions[i].Regions[j].ID < td.Partitions[i].Regions[k].ID
		})
	}

	for i := 0; i < len(td.Partitions); i++ {
		sort.SliceStable(td.Partitions[i].Services, func(j, k int) bool {
			return td.Partitions[i].Services[j].ID < td.Partitions[i].Services[k].ID
		})
	}

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("endpoints", tmpl, td, templateFuncMap); err != nil {
		g.Fatalf("error generating endpoint resolver: %s", err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

func readHTTPJSON(url string, to any) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return decodeFromReader(r.Body, to)
}

func decodeFromReader(r io.Reader, to any) error {
	dec := json.NewDecoder(r)

	for {
		if err := dec.Decode(to); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

//go:embed output.go.gtpl
var tmpl string
