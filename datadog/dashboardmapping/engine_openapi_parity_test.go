package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestOpenAPIParityListConstraints(t *testing.T) {
	computeSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:   "compute",
		Type:     TypeBlockList,
		MinItems: 1,
		MaxItems: 5,
		Children: []FieldSpec{{
			HCLKey:   "aggregation",
			Type:     TypeString,
			Required: true,
		}},
	})
	if computeSchema.MinItems != 1 || computeSchema.MaxItems != 5 {
		t.Fatalf("block list constraints were not registered: %#v", computeSchema)
	}

	listStreamSchema := FieldSpecsToSDKv2Schema(listStreamQueryFields)
	if listStreamSchema["group_by"].MaxItems != 4 || listStreamSchema["compute"].MinItems != 1 || listStreamSchema["compute"].MaxItems != 5 {
		t.Fatalf("list stream constraints do not match OpenAPI: %#v", listStreamSchema)
	}

	statesSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:      "states",
		Type:        TypeStringList,
		ValidValues: []string{"OPEN", "RESOLVED"},
	})
	elem, ok := statesSchema.Elem.(*schema.Schema)
	if !ok || elem.ValidateDiagFunc == nil {
		t.Fatalf("string-list enum validation was not registered: %#v", statesSchema.Elem)
	}
}

func TestOpenAPIParityQueryValueComparisonRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"query_value_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"q": "avg:system.cpu.user{*}",
				"comparison": []interface{}{map[string]interface{}{
					"type":           "absolute",
					"directionality": "neutral",
					"duration": []interface{}{map[string]interface{}{
						"type": "custom_timeframe",
						"custom_timeframe": []interface{}{map[string]interface{}{
							"from": 1779290190000,
							"to":   1779894990000,
						}},
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	comparison := request["comparison"].(map[string]interface{})
	if comparison["type"] != "absolute" || comparison["directionality"] != "neutral" {
		t.Fatalf("comparison defaults were not serialized: %#v", comparison)
	}
	duration := comparison["duration"].(map[string]interface{})
	custom := duration["custom_timeframe"].(map[string]interface{})
	if custom["from"] != 1779290190000 || custom["to"] != 1779894990000 {
		t.Fatalf("custom comparison timeframe was not serialized: %#v", custom)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["query_value_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	if _, ok := flatRequest["comparison"]; !ok {
		t.Fatalf("comparison was not restored during flatten: %#v", flatRequest)
	}
}

func TestOpenAPIParityListStreamQueryFieldsRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"list_stream_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"response_format": "event_list",
				"columns": []interface{}{map[string]interface{}{
					"field": "timestamp",
					"width": "auto",
				}},
				"query": []interface{}{map[string]interface{}{
					"data_source":      "issue_stream",
					"query_string":     "service:web-store",
					"states":           []interface{}{"OPEN", "ACKNOWLEDGED"},
					"assignee_uuids":   []interface{}{"user-uuid"},
					"suspected_causes": []interface{}{"deployment"},
					"team_handles":     []interface{}{"team-platform"},
					"persona":          "backend",
					"compute": []interface{}{map[string]interface{}{
						"aggregation": "count",
						"facet":       "resource_name",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	query := request["query"].(map[string]interface{})
	if query["persona"] != "backend" {
		t.Fatalf("list stream persona was not serialized: %#v", query)
	}
	if len(query["compute"].([]interface{})) != 1 {
		t.Fatalf("list stream compute was not serialized: %#v", query)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["list_stream_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatQuery := flatRequest["query"].([]interface{})[0].(map[string]interface{})
	if flatQuery["persona"] != "backend" {
		t.Fatalf("list stream fields were not restored during flatten: %#v", flatQuery)
	}
}

func TestOpenAPIParitySankeyAudienceFieldsRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"sankey_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"rum_request": []interface{}{map[string]interface{}{
					"query": []interface{}{map[string]interface{}{
						"data_source":  "product_analytics",
						"query_string": "@type:view",
						"mode":         "source",
						"audience_filters": []interface{}{map[string]interface{}{
							"user": []interface{}{map[string]interface{}{
								"name":  "buyers",
								"query": "@usr.plan:pro",
							}},
							"filter_condition": "users",
						}},
						"occurrences": []interface{}{map[string]interface{}{
							"operator": "gt",
							"value":    "2",
						}},
						"join_keys": []interface{}{map[string]interface{}{
							"primary":   "session.id",
							"secondary": []interface{}{"usr.id"},
						}},
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	request := definition["requests"].([]interface{})[0].(map[string]interface{})
	query := request["query"].(map[string]interface{})
	if request["request_type"] != "sankey" {
		t.Fatalf("sankey discriminator was not serialized: %#v", request)
	}
	if query["join_keys"].(map[string]interface{})["primary"] != "session.id" {
		t.Fatalf("sankey join keys were not serialized: %#v", query)
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["sankey_definition"].([]interface{})[0].(map[string]interface{})
	flatRequest := flatDefinition["request"].([]interface{})[0].(map[string]interface{})
	flatRum := flatRequest["rum_request"].([]interface{})[0].(map[string]interface{})
	flatQuery := flatRum["query"].([]interface{})[0].(map[string]interface{})
	if _, ok := flatQuery["audience_filters"]; !ok {
		t.Fatalf("sankey audience filters were not restored during flatten: %#v", flatQuery)
	}
}

func TestOpenAPIParityQueryTableSortBuildsForFormulaAndLegacyRequests(t *testing.T) {
	requests := []interface{}{
		map[string]interface{}{
			"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
			"query": []interface{}{map[string]interface{}{
				"metric_query": []interface{}{map[string]interface{}{
					"data_source": "metrics",
					"name":        "query1",
					"query":       "avg:system.cpu.user{*}",
				}},
			}},
			"sort": []interface{}{map[string]interface{}{
				"order_by": []interface{}{map[string]interface{}{
					"formula_sort": []interface{}{map[string]interface{}{"index": 1, "order": "desc"}},
				}},
			}},
		},
		map[string]interface{}{
			"q": "avg:system.mem.used{*}",
			"sort": []interface{}{map[string]interface{}{
				"order_by": []interface{}{map[string]interface{}{
					"group_sort": []interface{}{map[string]interface{}{"name": "host", "order": "asc"}},
				}},
			}},
		},
	}
	widget := map[string]interface{}{
		"query_table_definition": []interface{}{map[string]interface{}{"request": requests}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	builtRequests := definition["requests"].([]interface{})
	for i, request := range builtRequests {
		if _, ok := request.(map[string]interface{})["sort"]; !ok {
			t.Fatalf("sort missing from request %d: %#v", i, request)
		}
	}

	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["query_table_definition"].([]interface{})[0].(map[string]interface{})
	flatRequests := flatDefinition["request"].([]interface{})
	for i, request := range flatRequests {
		if _, ok := request.(map[string]interface{})["sort"]; !ok {
			t.Fatalf("sort missing after flatten for request %d: %#v", i, request)
		}
	}
}

func TestOpenAPIParityApmMetricsFormulaAndDistributionHistogram(t *testing.T) {
	apmMetricsQuery := map[string]interface{}{
		"data_source":    "apm_metrics",
		"name":           "query1",
		"stat":           "hits",
		"service":        "web-store",
		"operation_name": "web.request",
		"span_kind":      "server",
	}
	formulaWidget := map[string]interface{}{
		"query_value_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"formula": []interface{}{map[string]interface{}{"formula_expression": "query1"}},
				"query": []interface{}{map[string]interface{}{
					"apm_metrics_query": []interface{}{apmMetricsQuery},
				}},
			}},
		}},
	}
	builtFormula := BuildWidgetEngineJSONFromMap(formulaWidget)
	formulaDefinition := builtFormula["definition"].(map[string]interface{})
	formulaRequest := formulaDefinition["requests"].([]interface{})[0].(map[string]interface{})
	builtQuery := formulaRequest["queries"].([]interface{})[0].(map[string]interface{})
	if builtQuery["data_source"] != "apm_metrics" {
		t.Fatalf("APM metrics query was not serialized: %#v", builtQuery)
	}

	distributionWidget := map[string]interface{}{
		"distribution_definition": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{
				"request_type": "histogram",
				"histogram_query": []interface{}{map[string]interface{}{
					"apm_metrics_query": []interface{}{apmMetricsQuery},
				}},
			}},
		}},
	}
	builtDistribution := BuildWidgetEngineJSONFromMap(distributionWidget)
	distributionDefinition := builtDistribution["definition"].(map[string]interface{})
	distributionRequest := distributionDefinition["requests"].([]interface{})[0].(map[string]interface{})
	histogramQuery := distributionRequest["query"].(map[string]interface{})
	if distributionRequest["request_type"] != "histogram" || histogramQuery["data_source"] != "apm_metrics" {
		t.Fatalf("APM metrics histogram query was not serialized: %#v", distributionRequest)
	}
}

func TestOpenAPIParityFunnelGroupedDisplayRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"funnel_definition": []interface{}{map[string]interface{}{
			"grouped_display": "side_by_side",
			"request": []interface{}{map[string]interface{}{
				"query": []interface{}{map[string]interface{}{
					"data_source":  "rum",
					"query_string": "@type:view",
					"step": []interface{}{map[string]interface{}{
						"facet": "@view.name",
						"value": "/home",
					}},
				}},
			}},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	if definition["grouped_display"] != "side_by_side" {
		t.Fatalf("funnel grouped_display was not serialized: %#v", definition)
	}
	flattened, _ := FlattenWidgetEngineJSON(built)
	flatDefinition := flattened["funnel_definition"].([]interface{})[0].(map[string]interface{})
	if flatDefinition["grouped_display"] != "side_by_side" {
		t.Fatalf("funnel grouped_display was not restored: %#v", flatDefinition)
	}
}
