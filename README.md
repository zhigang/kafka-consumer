# Kafka Consumer

Api extensions for [Sarama Cluster](https://github.com/bsm/sarama-cluster), the Go client library for Apache Kafka 0.9 (and later).

## More Documentation

Sarama dluster documentation and example are available via godoc at <http://godoc.org/github.com/bsm/sarama-cluster>

Kafka quick start documentation at <http://kafka.apache.org/quickstart>

## Docker Command

```bash
docker run -d -p 3000:3000 -e KAFKA_BROKERS="kafka:9092" --name kafka-consumer siriuszg/kafka-consumer:TAG
```

## Http Request

```bash
curl 'http://localhost:3000/v1/consumer?topics=test&initial=new&count=10'
```

### Query Param

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
