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
