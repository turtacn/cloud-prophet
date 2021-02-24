# 把kube-scheduler改造成通用调度器

利用kubernetes调度器框架，扩展支持jvirt调度场景，重用`node`和`pod`的抽象概念，主打资源调度。

## 原始特点 

插件和扩展器来自kube-schedule framework, 原生的可扩展点

### 插件化(Plugins)

很多可扩展的步骤可以通过插件的方式集成到主调度Pipeline中，按照场景集成，意识是给定一个场景，可以在局部既支持过滤，又支持打分还支持其他的附件调度计算。

框架只负责主体流程，各个流程中的步骤次序、开关、权重都可以通过`profile`来配置，可以支持不同场景不同的调度。


|步骤|功能|
| --  | -- |
|prefilter|初粒度过滤，布隆过滤器，快速排除大量不匹配节点|
|filter | 精过滤 |
|postfilter| 代价，日志，统计  |
|reserved | 内存预占，调度周期结束|
|permit| 检查候选节点的实时状态 |
|prescore| 预打分 |
|score | 打分|
|postscore| 规格化打分 |
|prebind| 预绑定， 内存预占 |
|bind| 持久化入库绑定 |
|postbind| 代价，日志，统计信息|

#### 主要接口模型（差系统架构图）

```text 
type WaitingPod interface {}
type Plugin interface {}
type QueueSortPlugin interface {}
type PreFilterExtensions interface {}
type PreFilterPlugin interface {}
type FilterPlugin interface {}
type PostFilterPlugin interface {}
type PreScorePlugin interface {}
type ScoreExtensions interface {}
type ScorePlugin interface {}
type ReservePlugin interface {}
type PreBindPlugin interface {}
type PostBindPlugin interface {}
type PermitPlugin interface {}
type BindPlugin interface {}
type Framework interface {}
type FrameworkHandle interface {}
type PreemptHandle interface {}
type PodNominator interface {}
type PluginsRunner interface {}
type Extender interface {}
type ScheduleAlgorithm interface{}
type Binder interface {}
type Candidate interface {}
type StateData interface {}
type Framework interface {}
type FrameworkHandle interface {}
type PreemptHandle interface {}
type PodNominator interface {}
type PluginsRunner interface {}
type NodeInformLister interface {}
type SharedLister interface {}
type Cache interface {}
type SchedulingQueue interface {}
```


### 扩展

`Extender`可以替换接管上面步骤中的任何默认插件，支持外部二进制(IPC)的方式影响、扩展资源调度器本身，根据配置旁路默认插件过程。

## 扩展特性

### 启发式调度(Heuristics)

- 多目标
    - 能耗（Energy Consumption）
    - 资源利用率（Resource Utility）
    - 服务质量保障（SLA&QoS）   
    
- 限制条件
    - 标签计算扩大调度空间
    - 风险控制
        - 并发的突增的负载 
        - 资源瓶颈转移   
        
### 预测（Prediction）

通常各个维度的资源利用率的预测都是有用的，都是基于时序数据的分析, 相关动作CGroup封装。

- 预测资源利用率上限
    - CPU
    - Net I/O,

### 自调整

输出集群状态，尤其是与调度相关的资源布局状态，通过指标为外部系统提供更多的智能基础数据。 
 

## 更新所有代码
```shell
 find  . -name  '*.go'  -type f  -exec   sed  -i  '1i\\/\/'   {}  \;
```



## Add List-Watch to DBMS

```text
https://github.com/pingcap/tidb-binlog
``` 


