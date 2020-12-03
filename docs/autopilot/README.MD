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

- 关系

    * 实例有没有关联（直接，规格间接）VPA
    
    * VPA有无推荐，需要不需要推荐
    
    * 实例对于VPA的更新模式，自动还是关闭
    
    * 实际资源Scale操作模式，直接、渐进。。。
    
    * 集群状态就是所有实例状态的集合


- 实例的聚合状态




- 集群状态

### routines

## Updater
