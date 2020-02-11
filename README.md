# Kafka Consumer

Api extensions for [Sarama](https://github.com/Shopify/sarama), Sarama is an MIT-licensed Go client library for Apache Kafka version 0.8 (and later).

## More Documentation

Kafka quick start documentation at [kafka.apache.org](http://kafka.apache.org/quickstart)

## Docker Command

```bash
docker run -d -p 3000:3000 --restart=always -e KAFKA_BROKERS="kafka:9092" --name kafka-consumer siriuszg/kafka-consumer:TAG
```

* set environment ENABLE_PRODUCER="true" bind producer api
  * auto send message to topic
  * just for test

## Http Request

### Producer

```bash
curl 'http://localhost:3000/v1/producer?topic=test&count=1'
```

### Consumer

```bash
curl 'http://localhost:3000/v1/consumer?topics=test&initial=new&count=10'
```

### Consumer Query Param

* topics
  * REQUIRED
  * type: string
  * kafka topic name, multiple topic split by ','
* initial
  * REQUIRED
  * type: string
  * value: 'old' or 'new'
  * kafka message initial offset from oldest or newest
* count
  * REQUIRED
  * type: int
  * kafka message count
* group
  * type: string
  * kafka consumer group
* filter
  * type: string
  * only return kafka message contains this value
* commit
  * type: string
  * value: '1' or '0'
  * set '1' is kafka consumer mark offset

## Console Color

* 格式：
  * `\033[显示方式;前景色;背景色m`
  * `\x1b[显示方式;前景色%s\x1b[0m`
* 说明：

    |前景色|背景色|颜色|
    |---|---|---|
    |30|40|黑色|
    |31|41|红色|
    |32|42|绿色|
    |33|43|黃色|
    |34|44|蓝色|
    |35|45|紫红色|
    |36|46|青蓝色|
    |37|47|白色|

    |显示方式|意义|
    |---|---|
    |0|终端默认设置|
    |1|高亮显示|
    |4|使用下划线|
    |5|闪烁|
    |7|反白显示|
    |8|不可见|
