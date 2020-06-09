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
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 0
        },
        "id": 12,
        "panels": [],
        "repeat": null,
        "title": "CPU Usage",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "fillGradient": 0,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 1
        },
        "id": 1,
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {
          "dataLinks": []
        },
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeRegions": [],
        "timeShift": null,
        "title": "CPU Usage",
        "tooltip": {
          "shared": false,
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
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ],
        "yaxis": {
          "align": false,
          "alignLevel": null
        }
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 8
        },
        "id": 13,
        "panels": [],
        "repeat": null,
        "title": "CPU Quota",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "columns": [],
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 1,
        "fontSize": "100%",
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 9
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
        "nullPointMode": "null as zero",
        "options": {},
        "pageSize": null,
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "scroll": true,
        "seriesOverrides": [],
        "showHeader": true,
        "sort": {
          "col": 0,
          "desc": true
        },
        "spaceLength": 10,
        "stack": false,
        "steppedLine": false,
        "styles": [
          {
            "alias": "Time",
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "pattern": "Time",
            "type": "hidden"
          },
          {
            "alias": "CPU Usage",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #A",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "CPU Requests",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #B",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "CPU Requests %",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #C",
            "thresholds": [],
            "type": "number",
            "unit": "percentunit"
          },
          {
            "alias": "CPU Limits",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #D",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "CPU Limits %",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #E",
            "thresholds": [],
            "type": "number",
            "unit": "percentunit"
          },
          {
            "alias": "Pod",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": true,
            "linkTooltip": "Drill down",
            "linkUrl": "./d/{{ .Namespace }}/{{ .Namespace }}-3scale-kubernetes-compute-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-namespace=$namespace&var-pod=$__cell",
            "pattern": "pod",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "pattern": "/.*/",
            "thresholds": [],
            "type": "string",
            "unit": "short"
          }
        ],
        "targets": [
          {
            "expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "A",
            "step": 10
          },
          {
            "expr": "sum(kube_pod_container_resource_requests_cpu_cores{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "B",
            "step": 10
          },
          {
            "expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod) / sum(kube_pod_container_resource_requests_cpu_cores{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "C",
            "step": 10
          },
          {
            "expr": "sum(kube_pod_container_resource_limits_cpu_cores{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "D",
            "step": 10
          },
          {
            "expr": "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod) / sum(kube_pod_container_resource_limits_cpu_cores{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "E",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "CPU Quota",
        "tooltip": {
          "shared": false,
          "sort": 0,
          "value_type": "individual"
        },
        "transform": "table",
        "type": "table",
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
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 16
        },
        "id": 14,
        "panels": [],
        "repeat": null,
        "title": "Memory Usage",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "fillGradient": 0,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 17
        },
        "id": 3,
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {
          "dataLinks": []
        },
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"$namespace\", container!=\"\"}) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeRegions": [],
        "timeShift": null,
        "title": "Memory Usage (w/o cache)",
        "tooltip": {
          "shared": false,
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
            "format": "bytes",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ],
        "yaxis": {
          "align": false,
          "alignLevel": null
        }
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 24
        },
        "id": 15,
        "panels": [],
        "repeat": null,
        "title": "Memory Quota",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "columns": [],
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 1,
        "fontSize": "100%",
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 25
        },
        "id": 4,
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
        "nullPointMode": "null as zero",
        "options": {},
        "pageSize": null,
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "scroll": true,
        "seriesOverrides": [],
        "showHeader": true,
        "sort": {
          "col": 0,
          "desc": true
        },
        "spaceLength": 10,
        "stack": false,
        "steppedLine": false,
        "styles": [
          {
            "alias": "Time",
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "pattern": "Time",
            "type": "hidden"
          },
          {
            "alias": "Memory Usage",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #A",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Memory Requests",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #B",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Memory Requests %",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #C",
            "thresholds": [],
            "type": "number",
            "unit": "percentunit"
          },
          {
            "alias": "Memory Limits",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #D",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Memory Limits %",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #E",
            "thresholds": [],
            "type": "number",
            "unit": "percentunit"
          },
          {
            "alias": "Memory Usage (RSS)",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #F",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Memory Usage (Cache)",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #G",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Memory Usage (Swap)",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #H",
            "thresholds": [],
            "type": "number",
            "unit": "bytes"
          },
          {
            "alias": "Pod",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": true,
            "linkTooltip": "Drill down",
            "linkUrl": "./d/{{ .Namespace }}/{{ .Namespace }}-3scale-kubernetes-compute-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-namespace=$namespace&var-pod=$__cell",
            "pattern": "pod",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "pattern": "/.*/",
            "thresholds": [],
            "type": "string",
            "unit": "short"
          }
        ],
        "targets": [
          {
            "expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "A",
            "step": 10
          },
          {
            "expr": "sum(kube_pod_container_resource_requests_memory_bytes{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "B",
            "step": 10
          },
          {
            "expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod) / sum(kube_pod_container_resource_requests_memory_bytes{namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "C",
            "step": 10
          },
          {
            "expr": "sum(kube_pod_container_resource_limits_memory_bytes{cluster=\"$cluster\", namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "D",
            "step": 10
          },
          {
            "expr": "sum(container_memory_working_set_bytes{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod) / sum(kube_pod_container_resource_limits_memory_bytes{namespace=\"$namespace\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "E",
            "step": 10
          },
          {
            "expr": "sum(container_memory_rss{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "F",
            "step": 10
          },
          {
            "expr": "sum(container_memory_cache{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "G",
            "step": 10
          },
          {
            "expr": "sum(container_memory_swap{cluster=\"$cluster\", namespace=\"$namespace\",container!=\"\"}) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "H",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Memory Quota",
        "tooltip": {
          "shared": false,
          "sort": 0,
          "value_type": "individual"
        },
        "transform": "table",
        "type": "table",
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
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 32
        },
        "id": 16,
        "panels": [],
        "repeat": null,
        "title": "Network",
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
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 33
        },
        "id": 5,
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
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": false,
        "steppedLine": false,
        "styles": [
          {
            "alias": "Time",
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "pattern": "Time",
            "type": "hidden"
          },
          {
            "alias": "Current Receive Bandwidth",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #A",
            "thresholds": [],
            "type": "number",
            "unit": "Bps"
          },
          {
            "alias": "Current Transmit Bandwidth",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #B",
            "thresholds": [],
            "type": "number",
            "unit": "Bps"
          },
          {
            "alias": "Rate of Received Packets",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #C",
            "thresholds": [],
            "type": "number",
            "unit": "pps"
          },
          {
            "alias": "Rate of Transmitted Packets",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #D",
            "thresholds": [],
            "type": "number",
            "unit": "pps"
          },
          {
            "alias": "Rate of Received Packets Dropped",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #E",
            "thresholds": [],
            "type": "number",
            "unit": "pps"
          },
          {
            "alias": "Rate of Transmitted Packets Dropped",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": false,
            "linkTooltip": "Drill down",
            "linkUrl": "",
            "pattern": "Value #F",
            "thresholds": [],
            "type": "number",
            "unit": "pps"
          },
          {
            "alias": "Pod",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "link": true,
            "linkTooltip": "Drill down to pods",
            "linkUrl": "./d/{{ .Namespace }}/{{ .Namespace }}-3scale-kubernetes-compute-resources-pod?var-datasource=$datasource&var-cluster=$cluster&var-namespace=$namespace&var-pod=$__cell",
            "pattern": "pod",
            "thresholds": [],
            "type": "number",
            "unit": "short"
          },
          {
            "alias": "",
            "colorMode": null,
            "colors": [],
            "dateFormat": "YYYY-MM-DD HH:mm:ss",
            "decimals": 2,
            "pattern": "/.*/",
            "thresholds": [],
            "type": "string",
            "unit": "short"
          }
        ],
        "targets": [
          {
            "expr": "sum(irate(container_network_receive_bytes_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "A",
            "step": 10
          },
          {
            "expr": "sum(irate(container_network_transmit_bytes_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "B",
            "step": 10
          },
          {
            "expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "C",
            "step": 10
          },
          {
            "expr": "sum(irate(container_network_transmit_packets_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "D",
            "step": 10
          },
          {
            "expr": "sum(irate(container_network_receive_packets_dropped_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "E",
            "step": 10
          },
          {
            "expr": "sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "table",
            "instant": true,
            "intervalFactor": 2,
            "legendFormat": "",
            "refId": "F",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Current Network Usage",
        "tooltip": {
          "shared": false,
          "sort": 0,
          "value_type": "individual"
        },
        "transform": "table",
        "type": "table",
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
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 40
        },
        "id": 17,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 41
        },
        "id": 6,
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_receive_bytes_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Receive Bandwidth",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 48
        },
        "id": 18,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 49
        },
        "id": 7,
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_transmit_bytes_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Transmit Bandwidth",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 56
        },
        "id": 19,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 57
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Rate of Received Packets",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 64
        },
        "id": 20,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 65
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_receive_packets_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Rate of Transmitted Packets",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 72
        },
        "id": 21,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 73
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_receive_packets_dropped_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Rate of Received Packets Dropped",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      },
      {
        "collapsed": false,
        "gridPos": {
          "h": 1,
          "w": 24,
          "x": 0,
          "y": 80
        },
        "id": 22,
        "panels": [],
        "repeat": null,
        "title": "Network",
        "type": "row"
      },
      {
        "aliasColors": {},
        "bars": false,
        "dashLength": 10,
        "dashes": false,
        "datasource": "$datasource",
        "fill": 10,
        "gridPos": {
          "h": 7,
          "w": 24,
          "x": 0,
          "y": 81
        },
        "id": 11,
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
        "linewidth": 0,
        "links": [],
        "nullPointMode": "null as zero",
        "options": {},
        "percentage": false,
        "pointradius": 5,
        "points": false,
        "renderer": "flot",
        "seriesOverrides": [],
        "spaceLength": 10,
        "stack": true,
        "steppedLine": false,
        "targets": [
          {
            "expr": "sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\"$namespace\"}[$interval])) by (pod)",
            "format": "time_series",
            "intervalFactor": 2,
            "legendFormat": "{{`{{pod}}`}}",
            "legendLink": null,
            "refId": "A",
            "step": 10
          }
        ],
        "thresholds": [],
        "timeFrom": null,
        "timeShift": null,
        "title": "Rate of Transmitted Packets Dropped",
        "tooltip": {
          "shared": false,
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
            "format": "Bps",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": 0,
            "show": true
          },
          {
            "format": "short",
            "label": null,
            "logBase": 1,
            "max": null,
            "min": null,
            "show": false
          }
        ]
      }
    ],
    "refresh": "10s",
    "schemaVersion": 18,
    "style": "dark",
    "tags": [
      "kubernetes-mixin",
      "3scale"
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
          "datasource": "$datasource",
          "definition": "",
          "hide": 2,
          "includeAll": false,
          "label": "cluster",
          "multi": false,
          "name": "cluster",
          "options": [],
          "query": "label_values(kube_pod_info, cluster)",
          "refresh": 1,
          "regex": "",
          "skipUrlSync": false,
          "sort": 2,
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
          "auto": false,
          "auto_count": 30,
          "auto_min": "10s",
          "datasource": "prometheus",
          "hide": 2,
          "includeAll": false,
          "label": null,
          "multi": false,
          "name": "interval",
          "options": [
            {
              "selected": true,
              "text": "4h",
              "value": "4h"
            }
          ],
          "query": "4h",
          "refresh": 2,
          "regex": "",
          "skipUrlSync": false,
          "sort": 1,
          "tagValuesQuery": "",
          "tags": [],
          "tagsQuery": "",
          "type": "interval",
          "useTags": false
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
    "title": "{{ .Namespace }} / 3scale / Kubernetes / Compute Resources / Namespace (Pods)",
    "version": 1
}