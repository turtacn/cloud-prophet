# A Universal Resource Scheduler


A light-weight, universal resource scheduler for resource orchestrator systems. It is derived from kubernetes scheduler framwork, and adopt the popular concept of "node" and "pod". The pod here has no containers belongs to, just the concept of allocation of resource.

## Feature 

** Plugins

Many extensible steps could be a plugin step integrated into the schedule main pipeline. Each plugin was defined as an interface. Given a processing context, we could package all the steps in one process logic unit. 



|Steps|Functions|
| --  | -- |
|prefilter|初粒度过滤，布隆过滤器，快速排出大量不匹配节点|
|filter | 精过滤 |
|postfilter| 代价，日志，统计  |
|reserved | 内存预占，调度周期结束|
|permit| 检查候选节点的实时状态 |
|prescore| 预打分 |
|score | 打分|
|postscore||
|prebind||
|bind||
|postbind||


** Extension

The extender could replace any plugin of the main schedule plan. It would be standalone binary program, acting as callee service, embedded in those schedule pipeline. If some external definition and configuration for this scheduler is given,
it the bypass the default process unit predefined in our schedule framwork. 

** Heuristics

- Multiple Objectives 
    - Energy Consumption
    - Resource Utility
    - QoS Gurantee & SLA Permission
    - 
   
    
- With Constrains
    - Schedulable Space limitation for many labels
    - Risks when high utility
        - Burst workload 
        - Other resource bottlenecks 
        - Hard to control  
    - Topology and plan 
    

** Prediction

We use predicted resource utility, here we call predicted resource limit upper. 


** Feasible 
 

 


