{
    "annotations": {
        "list": [
            {
                "builtIn": 1,
                "datasource": "-- Grafana --",
                "enable": true,
                "hide": true,
                "iconColor": "rgba(0, 211, 255, 1)",
                "name": "Annotations & Alerts",
                "type": "dashboard"
            }
        ]
    },
    "editable": true,
    "gnetId": null,
    "graphTooltip": 0,
    "links": [],
    "panels": [
        {
            "collapsed": false,
            "gridPos": {
                "h": 1,
                "w": 24,
                "x": 0,
                "y": 0
            },
            "id": 20,
            "panels": [],
            "repeat": "service_id",
            "title": "$service_id - $service_name",
            "type": "row"
        },
        {
            "aliasColors": {},
            "bars": false,
            "dashLength": 10,
            "dashes": false,
            "datasource": "$datasource",
            "fill": 1,
            "gridPos": {
                "h": 10,
                "w": 12,
                "x": 0,
                "y": 1
            },
            "id": 60,
            "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
            },
            "lines": true,
            "linewidth": 1,
            "links": [],
            "nullPointMode": "null",
            "options": {},
            "percentage": false,
            "pointradius": 2,
            "points": false,
            "renderer": "flot",
            "seriesOverrides": [],
            "spaceLength": 10,
            "stack": false,
            "steppedLine": false,
            "targets": [
                {
                    "expr": "sum(rate(upstream_status{namespace='$namespace', pod=~'apicast-$env.*', service_id='$service_id'}[1m])) by (status)",
                    "format": "time_series",
                    "intervalFactor": 1,
                    "legendFormat": "{{`{{status}}`}}",
                    "refId": "A"
                }
            ],
            "thresholds": [],
            "timeFrom": null,
            "timeRegions": [],
            "timeShift": null,
            "title": "Upstream HTTP status codes (per second rate)",
            "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
            },
            "type": "graph",
            "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
            },
            "yaxes": [
                {
                    "decimals": 0,
                    "format": "short",
                    "label": null,
                    "logBase": 1,
                    "max": null,
                    "min": null,
                    "show": true
                },
                {
                    "format": "reqps",
                    "label": null,
                    "logBase": 1,
                    "max": null,
                    "min": null,
                    "show": true
                }
            ],
            "yaxis": {
                "align": false,
                "alignLevel": null
            }
        },
        {
            "aliasColors": {},
            "bars": false,
            "dashLength": 10,
            "dashes": false,
            "fill": 5,
            "gridPos": {
                "h": 10,
                "w": 12,
                "x": 12,
                "y": 1
            },
            "id": 63,
            "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
            },
            "lines": true,
            "linewidth": 1,
            "links": [],
            "nullPointMode": "null",
            "options": {},
            "percentage": false,
            "pointradius": 2,
            "points": false,
            "renderer": "flot",
            "seriesOverrides": [],
            "spaceLength": 10,
            "stack": false,
            "steppedLine": false,
            "targets": [
                {
                    "expr": "histogram_quantile($percentile/100, sum(rate(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env.*', service_id='$service_id'}[1m])) by (le))",
                    "format": "time_series",
                    "intervalFactor": 1,
                    "legendFormat": "Total request time",
                    "refId": "A"
                },
                {
                    "expr": "histogram_quantile($percentile/100, sum(rate(upstream_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env.*', service_id='$service_id'}[1m])) by (le))",
                    "format": "time_series",
                    "intervalFactor": 1,
                    "legendFormat": "Upstream request time",
                    "refId": "B"
                }
            ],
            "thresholds": [],
            "timeFrom": null,
            "timeRegions": [],
            "timeShift": null,
            "title": "Total request time vs upstream request time (${percentile}th percentile)",
            "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
            },
            "type": "graph",
            "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
            },
            "yaxes": [
                {
                    "decimals": null,
                    "format": "s",
                    "label": "",
                    "logBase": 1,
                    "max": null,
                    "min": null,
                    "show": true
                },
                {
                    "format": "short",
                    "label": null,
                    "logBase": 1,
                    "max": null,
                    "min": null,
                    "show": true
                }
            ],
            "yaxis": {
                "align": false,
                "alignLevel": null
            }
        },
        {
            "cards": {
                "cardPadding": null,
                "cardRound": null
            },
            "color": {
                "cardColor": "#73BF69",
                "colorScale": "sqrt",
                "colorScheme": "interpolateOranges",
                "exponent": 0.5,
                "mode": "opacity"
            },
            "dataFormat": "tsbuckets",
            "datasource": "$datasource",
            "gridPos": {
                "h": 10,
                "w": 12,
                "x": 0,
                "y": 11
            },
            "heatmap": {},
            "hideZeroBuckets": false,
            "highlightCards": true,
            "id": 58,
            "legend": {
                "show": true
            },
            "links": [],
            "options": {},
            "reverseYBuckets": false,
            "targets": [
                {
                    "expr": "sum(rate(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env-.*', service_id='$service_id'}[1m])) by (le)",
                    "format": "heatmap",
                    "instant": false,
                    "intervalFactor": 10,
                    "legendFormat": "{{`{{le}}`}}",
                    "refId": "A"
                }
            ],
            "timeFrom": null,
            "timeShift": null,
            "title": "Response time heatmap (by time bucket)",
            "tooltip": {
                "show": true,
                "showHistogram": true
            },
            "type": "heatmap",
            "xAxis": {
                "show": true
            },
            "xBucketNumber": null,
            "xBucketSize": null,
            "yAxis": {
                "decimals": 0,
                "format": "s",
                "logBase": 1,
                "max": null,
                "min": null,
                "show": true,
                "splitFactor": null
            },
            "yBucketBound": "auto",
            "yBucketNumber": null,
            "yBucketSize": null
        },
        {
            "cards": {
                "cardPadding": null,
                "cardRound": null
            },
            "color": {
                "cardColor": "#FADE2A",
                "colorScale": "sqrt",
                "colorScheme": "interpolateOranges",
                "exponent": 0.5,
                "mode": "opacity"
            },
            "dataFormat": "tsbuckets",
            "datasource": "$datasource",
            "gridPos": {
                "h": 10,
                "w": 12,
                "x": 12,
                "y": 11
            },
            "heatmap": {},
            "hideZeroBuckets": false,
            "highlightCards": true,
            "id": 61,
            "legend": {
                "show": true
            },
            "links": [],
            "options": {},
            "reverseYBuckets": false,
            "targets": [
                {
                    "expr": "sum(rate(upstream_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env.*', service_id='$service_id'}[1m])) by (le)",
                    "format": "heatmap",
                    "intervalFactor": 10,
                    "legendFormat": "{{`{{le}}`}}",
                    "refId": "A"
                }
            ],
            "timeFrom": null,
            "timeShift": null,
            "title": "Upstream response time heatmap (by time bucket)",
            "tooltip": {
                "show": true,
                "showHistogram": true
            },
            "type": "heatmap",
            "xAxis": {
                "show": true
            },
            "xBucketNumber": null,
            "xBucketSize": null,
            "yAxis": {
                "decimals": 0,
                "format": "s",
                "logBase": 1,
                "max": null,
                "min": null,
                "show": true,
                "splitFactor": null
            },
            "yBucketBound": "auto",
            "yBucketNumber": null,
            "yBucketSize": null
        }
    ],
    "refresh": "10s",
    "schemaVersion": 18,
    "style": "dark",
    "tags": [
        "3scale",
        "apicast"
    ],
    "templating": {
        "list": [
            {
                "hide": 0,
                "includeAll": false,
                "label": null,
                "multi": false,
                "name": "datasource",
                "options": [],
                "query": "prometheus",
                "refresh": 1,
                "regex": "",
                "skipUrlSync": false,
                "type": "datasource"
            },
            {
                "allValue": null,
                "current": {
                    "tags": [],
                    "text": "{{ .Namespace }}",
                    "value": "{{ .Namespace }}"
                },
                "hide": 0,
                "includeAll": false,
                "label": "namespace",
                "multi": false,
                "name": "namespace",
                "options": [
                    {
                        "selected": true,
                        "text": "{{ .Namespace }}",
                        "value": "{{ .Namespace }}"
                    }
                ],
                "query": "{{ .Namespace }}",
                "skipUrlSync": false,
                "type": "custom"
            },
            {
                "allValue": null,
                "current": {
                    "selected": false,
                    "tags": [],
                    "text": "production",
                    "value": "production"
                },
                "hide": 0,
                "includeAll": false,
                "label": "environment",
                "multi": false,
                "name": "env",
                "options": [
                    {
                        "selected": true,
                        "text": "production",
                        "value": "production"
                    },
                    {
                        "selected": false,
                        "text": "staging",
                        "value": "staging"
                    }
                ],
                "query": "production,staging",
                "skipUrlSync": false,
                "type": "custom"
            },
            {
                "allValue": null,
                "datasource": "$datasource",
                "definition": "label_values(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env.*'}, service_id)",
                "hide": 0,
                "includeAll": false,
                "label": "",
                "multi": true,
                "name": "service_id",
                "options": [],
                "query": "label_values(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env.*'}, service_id)",
                "refresh": 2,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
            },
            {
                "allValue": null,
                "current": {
                    "text": "api",
                    "value": "api"
                },
                "datasource": "$datasource",
                "definition": "label_values(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env-.*', service_id='$service_id'}, service_system_name)",
                "hide": 2,
                "includeAll": false,
                "label": "",
                "multi": true,
                "name": "service_name",
                "options": [
                    {
                        "selected": true,
                        "text": "api",
                        "value": "api"
                    }
                ],
                "query": "label_values(total_response_time_seconds_bucket{namespace='$namespace', pod=~'apicast-$env-.*', service_id='$service_id'}, service_system_name)",
                "refresh": 0,
                "regex": "",
                "skipUrlSync": false,
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
            },
            {
                "allValue": null,
                "current": {
                    "tags": [],
                    "text": "95",
                    "value": "95"
                },
                "hide": 0,
                "includeAll": false,
                "label": null,
                "multi": false,
                "name": "percentile",
                "options": [
                    {
                        "selected": false,
                        "text": "99",
                        "value": "99"
                    },
                    {
                        "selected": true,
                        "text": "95",
                        "value": "95"
                    },
                    {
                        "selected": false,
                        "text": "90",
                        "value": "90"
                    },
                    {
                        "selected": false,
                        "text": "80",
                        "value": "80"
                    },
                    {
                        "selected": false,
                        "text": "70",
                        "value": "70"
                    },
                    {
                        "selected": false,
                        "text": "50",
                        "value": "50"
                    }
                ],
                "query": "99, 95, 90, 80, 70, 50",
                "skipUrlSync": false,
                "type": "custom"
            }
        ]
    },
    "time": {
        "from": "now-1h",
        "to": "now"
    },
    "timepicker": {
        "refresh_intervals": [
            "5s",
            "10s",
            "30s",
            "1m",
            "5m",
            "15m",
            "30m",
            "1h",
            "2h",
            "1d"
        ],
        "time_options": [
            "5m",
            "15m",
            "1h",
            "6h",
            "12h",
            "24h",
            "2d",
            "7d",
            "30d"
        ]
    },
    "timezone": "",
    "title": "{{ .Namespace }} / 3scale / Apicast Services"
}