# 资源利用率预测

基于jdcloud线上真实数据，利用现有经典机器学习方法对资源利用率建模，分析集合维度、实例维度的资源利用率，并通过机器学习算法预测资源利用率的变化。

## 数据采集

### 集合数据

所有物理节点、实例在时间范围内的平均CPU/内存的平均、最大、top95、top99等指标分布

### 实例数据

所有物理节点/虚拟实例的CPU利用率、内存利用率、磁盘利用率、网络出口利用率、网络入口利用率

目前数据集为一周的采集数据，聚合采用max，降采用为60s-avg，从[HoraeDB API](https://cf.jd.com/pages/viewpage.action?pageId=188074271)抓取

## 模型训练

* COV

1D Convolutional Neural Network: 一维卷积神经网络

* GRU * 2

GRU（Gate Recurrent Unit）是循环神经网络（Recurrent Neural Network, RNN）的一种。和LSTM（Long-Short Term Memory）一样，也是为了解决长期记忆和反向传播中的梯度等问题而提出来的。
GRU和LSTM在很多情况下实际表现上相差无几，但比LSTM容易训练

* Dense * 2

密集特征(dense feature)

* Dropout

dropout 是指在深度学习网络的训练过程中，按照一定的概率将一部分神经网络单元暂时从网络中丢弃，相当于从原始的网络中找到一个更瘦的网络

* Fbprophet

facebook开源的的一个时间序列预测算法，能够几乎全自动地预测时间序列未来地走势。 它是基于时间序列分解和机器学习的拟合来做的，完全具备数理统计的基础。

<!--
|GRU+D+E50|||
|GRU+D+D+E50|||
|GRU+GRU+D+D+E50|||
|GRU+GRU+D+D+DROP+E50|||
|COV+GRU+GRU+D+D+E50|||
|LSTM|||
-->

|算法|RMSE|MSE|
|--|--|--|
|COV+GRU+GRU+D+D++DROP+E50| 0.02676973|0.00071662|
|fbprophet|||

### 训练误差

* 京东 vs 阿里

![loss_train](/experiments/jdcloud/loss_train.png)


![loss_train](/experiments/alibaba/loss_train.png)



## 预测

* 京东1 vs  阿里1

![jdpredict](/experiments/jdcloud/prediction.png)

![alipredict](/experiments/alibaba/prediction.png)


* 京东2 vs  阿里2

![jdpredict](/experiments/jdcloud/prediction_zoom1.png)

![alipredict](/experiments/alibaba/prediction_zoom1.png)



* 京东3 vs  阿里3

![jdpredict](/experiments/jdcloud/prediction_zoom2.png)

![alipredict](/experiments/alibaba/prediction_zoom2.png)


<!--
jdcloudiaas/turta:pydata
jdcloudiaas/turta:vim:4c448496813c:3:days:ago:2.77GB
jdcloudiaas/turta:7thvim:67b84bed9256:4:days:ago:1.53GB
jdcloudiaas/turta:govim:757fd68d57d5:7:days:ago:1.71GB
jdcloudiaas/turta:vimbase:6c77f6ac9c07:8:days:ago:1.56GB

url = https://github.com/NeuronEmpire/aliyun_schedule_semi.git
url = https://github.com/Ferdib-Al-Islam/fb-prophet-time-series-forcasting.git
url = https://github.com/chrislusf/gleam.git
url = https://github.com/sssssch/jupyter-examples.git
url = https://github.com/SOUP-KMITL/Thoth.git
url = https://github.com/gardener/remedy-controller.git
-->