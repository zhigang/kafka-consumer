# kafka consumer

## Howto
### Quick Start
```
docker run -d -p 3000:3000 --restart=always -e KAFKA_BROKERS="kafka:9092" --name kafka-consumer siriuszg/kafka-consumer:TAG
```