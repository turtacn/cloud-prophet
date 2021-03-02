# Implementation of google's autopilot

## 模型

延用kubernetes的资源层次模型：Cluster, Node, Pod, Container

资源管理的粒度是Pod

## 原理

基于历史数据预测（加权分位数估计）


## CommandLine

```text
autopilot --sample-second-interval 60 --target-cpu-percentile 0.8 --csv-file all-1604325990-1604995590-172.19.9.104-usage.csv --element-id 172.19.9.104 --cpu-histogram-decay-half-life 3h0m0s --recommendation-margin-fraction 1.6
```

