# Scheduling Internals in Kubernetes

## Project Overview:
Analyze the kubernetes controller framework that is implemented in the scheduler, build tests which can rapidly simulate how it interacts with etcd in a large cluster, and use those tests to determine if algorithmic improvements (i.e. bloom filters) can be used to speed up matching of pods -> nodes in a busy cluster.

Chance to deeply learn about scheduling, golang, kubernetes, etcd, consistency algorithms,the RAFT algorithm, bloom filters, and database n-ary indices. High profile pull requests into at least 3 projects will be possible: etcd, kubernetes, and willf's bloom filter.

0) Build scheduling stress tests inside of kubernetes unit testing library which measure the time to schedule 1000s of pods when node matching needs to occur using complex inequality/predicate logic (i.e. memory, cpu constraints, etc.). This is mostly kubernetes coding activity.

1) Research: a bloom filter with inequalities, e.g. this . Can it be used in the real world, when and when not? Implement it as an extension to https://github.com/willf/bloom . This is pure algorithmic research.

2) Analyze how the controller interacts with etcd, diagram it, and determine what happens when inequalities (less than, greater than) are queried. Look at the Matches interface and how it is triggered from the current scheduler. This is mostly pure golang+kube research: Diagramming and looking at the code path.

3) Analyze ETCD (this is the backbone of kubernetes: the golang equivalent of zookeeper: a k/v database that is strictly consistent) and determine how one can efficiently query inequalities in etcd. How range queries currently supported in ETCD, if so, how are they implemented, and could they be sped up? (e.g. this ) . Also, in general, how are range queries implemented in other K/V databases (SOLR, Cassandra,...).

4) Find where the kubernetes API uses ETCD, intercept all calls: What parts of etcd's API is it using/not-using. This can be done using anything from Wireshark just look at HTTP calls being made - to just reading through-analyzing of the code. Make a meaningfull histogram of etcd calls being made in a real world application (i.e. running density tests, or running other e2e tests).

## Overall goal

Determine wether or not (1) can be used to optimize the performance of queries inside of (2). Measure the difference using the results of (0). (3 and 4 play into this because they will allow us to know if we're on the right path for 1 and 2, or wether we should change direction).
