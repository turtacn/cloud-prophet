# 虚机动态平衡

## 算法
|||
|--|--|
|[SAT'17]|[Introducing Pareto Minimal Correction Subsets](https://link.springer.com/chapter/10.1007/978-3-319-66263-3_13)|
|[AAAI'18]|[Enhancing Constraint-Based Multi-objective Conbinatorial Optimization](papers/Enhancing%20Constraint-Based.pdf)|
|[IJCAI'18]|[Stratification for Constraint-Based Multi-Objective Combinatorial Optimization](papers/Stratification%20for%20Constraint-Based%20Multi-Objective%20Combinatorial%20Optimization.pdf)|
|[Heuristics'18]|[Virtual machine consolidation using constraint-based multi-objective optimization](https://link.springer.com/article/10.1007/s10732-018-9400-2)|
|[IJCAI'19]|[Integrating Pseudo-Boolean Constraint Reasoning in multi-objective evelutionary algorithm](papers/Integrating%20Pseudo-Boolean%20Constraint%20Reasoning%20in%20multi-objective%20evolutionary%20algorithms.pdf)|
|[VMPMBBO]|[Virtual machine consolidated placement based on multi-objective biogeography-based optimization](papers/Virtual%20machine%20consolidated%20placement%20based%20on%20multi-objective%20biogeography-based%20optimization.pdf)|

## 评估

### 测评数据格式

The format of the benchmarks is as follows:
- 1 line with number of servers (N)
- N lines with server descriptions: server id, CPU capacity, memory capacity, idle energy consumption and maximum energy consumption
- 1 line with number of virtual machines (M)
- M lines with virtual machine descriptions: job id, virtual machine index in job, CPU requirements, memory requirements, Boolean that is True if virtual machine cannot be co-located with other virtual machines in the same job (False otherwise); some virtual machines in some instances may include a comma separated list of server ids corresponding to servers where that virtual machine is not allowed to run, referred to as incompatible servers; these can be ignored using the **-ip** option when running vmalloc.jar
- 1 line with number of pre-existing mappings (K)
- K lines with mapping descriptions: job id, virtual machine index in job, server id

```
<number of servers (N)>
<server id 1> <cpu capacity 1> <memory capacity 1> <idle energy consumption 1> <maximum energy consumption 1>
    ...
<server id N> <cpu capacity N> <memory capacity N> <idle energy consumption N> <maximum energy consumption N>
<number of virtual machines (M)>
<job id 1> <virtual machine index 1> <cpu requirements 1> <memory requirements 1> <is anti co-locatable 1>
    ...
<job id M> <virtual machine index M> <cpu requirements M> <memory requirements M> <is anti co-locatable M>
<number of mappings (K)>
<mapping job id 1> <mapping virtual machine index 1> <mapping server id 1>
    ...
<mapping job id K> <mapping virtual machine index K> <mapping server id K>
```






