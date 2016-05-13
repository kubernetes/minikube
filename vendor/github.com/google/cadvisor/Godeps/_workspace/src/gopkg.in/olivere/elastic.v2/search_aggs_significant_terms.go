// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

// SignificantSignificantTermsAggregation is an aggregation that returns interesting
// or unusual occurrences of terms in a set.
// See: http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/search-aggregations-bucket-significantterms-aggregation.html
type SignificantTermsAggregation struct {
	field           string
	subAggregations map[string]Aggregation

	minDocCount      *int
	shardMinDocCount *int
	requiredSize     *int
	shardSize        *int
	filter           Filter
	executionHint    string
}

func NewSignificantTermsAggregation() SignificantTermsAggregation {
	a := SignificantTermsAggregation{
		subAggregations: make(map[string]Aggregation, 0),
	}
	return a
}

func (a SignificantTermsAggregation) Field(field string) SignificantTermsAggregation {
	a.field = field
	return a
}

func (a SignificantTermsAggregation) SubAggregation(name string, subAggregation Aggregation) SignificantTermsAggregation {
	a.subAggregations[name] = subAggregation
	return a
}

func (a SignificantTermsAggregation) MinDocCount(minDocCount int) SignificantTermsAggregation {
	a.minDocCount = &minDocCount
	return a
}

func (a SignificantTermsAggregation) ShardMinDocCount(shardMinDocCount int) SignificantTermsAggregation {
	a.shardMinDocCount = &shardMinDocCount
	return a
}

func (a SignificantTermsAggregation) RequiredSize(requiredSize int) SignificantTermsAggregation {
	a.requiredSize = &requiredSize
	return a
}

func (a SignificantTermsAggregation) ShardSize(shardSize int) SignificantTermsAggregation {
	a.shardSize = &shardSize
	return a
}

func (a SignificantTermsAggregation) BackgroundFilter(filter Filter) SignificantTermsAggregation {
	a.filter = filter
	return a
}

func (a SignificantTermsAggregation) ExecutionHint(hint string) SignificantTermsAggregation {
	a.executionHint = hint
	return a
}

func (a SignificantTermsAggregation) Source() interface{} {
	// Example:
	// {
	//     "query" : {
	//         "terms" : {"force" : [ "British Transport Police" ]}
	//     },
	//     "aggregations" : {
	//         "significantCrimeTypes" : {
	//             "significant_terms" : { "field" : "crime_type" }
	//         }
	//     }
	// }
	//
	// This method returns only the
	//   { "significant_terms" : { "field" : "crime_type" }
	// part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source["significant_terms"] = opts

	if a.field != "" {
		opts["field"] = a.field
	}
	if a.requiredSize != nil {
		opts["size"] = *a.requiredSize // not a typo!
	}
	if a.shardSize != nil {
		opts["shard_size"] = *a.shardSize
	}
	if a.minDocCount != nil {
		opts["min_doc_count"] = *a.minDocCount
	}
	if a.shardMinDocCount != nil {
		opts["shard_min_doc_count"] = *a.shardMinDocCount
	}
	if a.filter != nil {
		opts["background_filter"] = a.filter.Source()
	}
	if a.executionHint != "" {
		opts["execution_hint"] = a.executionHint
	}

	// AggregationBuilder (SubAggregations)
	if len(a.subAggregations) > 0 {
		aggsMap := make(map[string]interface{})
		source["aggregations"] = aggsMap
		for name, aggregate := range a.subAggregations {
			aggsMap[name] = aggregate.Source()
		}
	}

	return source
}
