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
  "id": 1,
  "links": [],
  "panels": [
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "id": 14,
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
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "authorize",
          "refId": "A"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "authrep",
          "refId": "B"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"report\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "report",
          "refId": "C"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize_oauth\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "authorize_oauth",
          "refId": "D"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep_oauth\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "authrep_oauth",
          "refId": "E"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Listener requests per second (by request type)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
      "alert": {
        "conditions": [
          {
            "evaluator": {
              "params": [
                100
              ],
              "type": "gt"
            },
            "operator": {
              "type": "and"
            },
            "query": {
              "params": [
                "C",
                "5m",
                "now"
              ]
            },
            "reducer": {
              "params": [],
              "type": "sum"
            },
            "type": "query"
          }
        ],
        "executionErrorState": "alerting",
        "for": "5m",
        "frequency": "1m",
        "handler": 1,
        "name": "Listener response codes per second alert",
        "noDataState": "no_data",
        "notifications": []
      },
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 0
      },
      "id": 12,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": true,
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
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",resp_code=\"2xx\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "2xx",
          "refId": "A"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",resp_code=\"403\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "403",
          "refId": "B"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",resp_code=\"404\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "404",
          "refId": "D"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",resp_code=\"409\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "409",
          "refId": "E"
        },
        {
          "expr": "sum(rate(apisonator_listener_response_codes{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",resp_code=\"5xx\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "5xx",
          "refId": "C"
        }
      ],
      "thresholds": [
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "gt",
          "value": 100
        }
      ],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Listener response codes per second",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 8
      },
      "id": 16,
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
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize\"}[1m])",
          "format": "heatmap",
          "interval": "",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Authorize response times (per time bucket)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#FADE2A",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 8
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 17,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Authorize response times (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 16
      },
      "id": 24,
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
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep\"}[1m])",
          "format": "heatmap",
          "interval": "",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Authrep response times (per time bucket)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#5794F2",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 16
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 25,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Authrep response times (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 24
      },
      "id": 20,
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
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"report\"}[1m])",
          "format": "heatmap",
          "interval": "",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Report response times (per time bucket)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#F2495C",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 24
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 23,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"report\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Report response times (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 32
      },
      "id": 22,
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
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize_oauth\"}[1m])",
          "format": "heatmap",
          "interval": "",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Authorize (oauth) response times (per time bucket)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#FF9830",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 32
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 21,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authorize_oauth\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Authorize (oauth) response times (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 40
      },
      "id": 18,
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
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep_oauth\"}[1m])",
          "format": "heatmap",
          "interval": "",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Authrep (oauth) response times (per time bucket)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#B877D9",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 40
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 19,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_listener_response_times_seconds_bucket{job=\"backend-listener-monitoring\",namespace=\"{{ .Namespace }}\",request_type=\"authrep_oauth\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Authrep (oauth) response times (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 0,
        "y": 48
      },
      "id": 2,
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
          "expr": "rate(apisonator_worker_job_count{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"ReportJob\"} [1m])",
          "format": "time_series",
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "Report jobs per second",
          "refId": "A"
        },
        {
          "expr": "rate(apisonator_worker_job_count{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"NotifyJob\"} [1m])",
          "format": "time_series",
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "Notify jobs per second",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Jobs per second",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 12,
        "y": 48
      },
      "id": 8,
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
          "expr": "apisonator_worker_job_count{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"ReportJob\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Report jobs",
          "refId": "A"
        },
        {
          "expr": "apisonator_worker_job_count{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"NotifyJob\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Notify jobs",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Total jobs processed",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
      "aliasColors": {},
      "bars": false,
      "cacheTimeout": null,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 0,
        "y": 57
      },
      "id": 9,
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
      "pluginVersion": "6.2.4",
      "pointradius": 2,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(apisonator_worker_job_runtime_seconds_bucket{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"ReportJob\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Report jobs runtime (seconds)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#5794F2",
        "colorScale": "sqrt",
        "colorScheme": "interpolateBlues",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 12,
        "y": 57
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 4,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_worker_job_runtime_seconds_bucket{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"ReportJob\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Report jobs runtime (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
      "aliasColors": {},
      "bars": false,
      "cacheTimeout": null,
      "dashLength": 10,
      "dashes": false,
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 0,
        "y": 66
      },
      "id": 10,
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
      "pluginVersion": "6.2.4",
      "pointradius": 2,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(apisonator_worker_job_runtime_seconds_bucket{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"NotifyJob\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Notify jobs runtime (seconds)",
      "tooltip": {
        "shared": true,
        "sort": 0,
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
          "format": "short",
          "label": null,
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
        "cardColor": "#b4ff00",
        "colorScale": "sqrt",
        "colorScheme": "interpolateGreens",
        "exponent": 0.5,
        "mode": "opacity"
      },
      "dataFormat": "tsbuckets",
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 12,
        "y": 66
      },
      "heatmap": {},
      "hideZeroBuckets": false,
      "highlightCards": true,
      "id": 6,
      "legend": {
        "show": false
      },
      "links": [],
      "options": {},
      "reverseYBuckets": false,
      "targets": [
        {
          "expr": "rate(apisonator_worker_job_runtime_seconds_bucket{job=\"backend-worker-monitoring\",namespace=\"{{ .Namespace }}\",type=\"NotifyJob\"}[1m])",
          "format": "heatmap",
          "interval": "1m",
          "intervalFactor": 10,
          "legendFormat": "{{`{{le}}`}}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Notify jobs runtime (seconds)",
      "tooltip": {
        "show": true,
        "showHistogram": false
      },
      "type": "heatmap",
      "xAxis": {
        "show": true
      },
      "xBucketNumber": null,
      "xBucketSize": null,
      "yAxis": {
        "decimals": null,
        "format": "short",
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
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-24h",
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
  "title": "{{ .Namespace }}/Apisonator",
  "version": 12058
}
