# VPA for engineering


![vpa](vpa.png)

## Recommender

### input 

- history 

从prometheus获取CPU、Memory历史指标，8d, 30s

- metrics 

获取实时实例的资源指标，数据成为快照信息，快照记录时间戳和时间窗口

- oom 

watch 实例 OOM事件，拿到OOM的实例映射

- spec

pod list


### checkpoint

为clusterState， 所有(实例, VPA)的信息提供CRUD操作，并持久化


### model

- 常量

    * OOMBumpUpRatio， 1.2

    * OOMMinBumpUp， 100MB

    * RecommendationMissingMaxDuration, 30分钟
    
    * DefaultMemoryAggregationCount, 8， 聚合历史8天的数据
    
    * DefaultMemoryAggregationInterval, 24*hours
    
    * DefaultHistogramBucketSizeGrowth, 0.05 保证每个桶比之前的大5%
    
    * DefaultMemoryHistogramDecayHalfLife， 24*hours 内存历史数据置信半衰期
    
    * DefaultCPUHistogramDecayHalfLife， 24*hours CPU历史数据置信半衰期
       
  
- 关系

    * 实例有没有关联（直接，规格间接）VPA
    
    * VPA有无推荐，需要不需要推荐
    
    * 实例对于VPA的更新模式，自动还是关闭
    
    * 实际资源Scale操作模式，直接、渐进。。。
    
    * 集群状态就是所有实例状态的集合


- 实例的聚合

聚合器提供增加采样，移除采样，输出推荐，输出更新模式


- 集群状态

实例及其状态的集合; VPA集合; 空VPA集合，无推荐; 观测VPA，和更新策略相关

### logic

- 预测器

各类估计器，通过实例的聚合计算资源的各类利用率估计

- 推荐引擎

一堆预测器的组合，包括算子和计算规则， 面向业务支持上下限、趋势


### routines

## Updater
