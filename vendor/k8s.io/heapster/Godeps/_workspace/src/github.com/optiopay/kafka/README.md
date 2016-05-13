[![Build Status](https://travis-ci.org/optiopay/kafka.svg?branch=master)](https://travis-ci.org/optiopay/kafka)
[![GoDoc](https://godoc.org/github.com/optiopay/kafka?status.png)](https://godoc.org/github.com/optiopay/kafka)

# Kafka

Kafka is Go client library for [Apache Kafka](https://kafka.apache.org/)
server, released under [MIT license](LICENSE]).

Kafka provides minimal abstraction over wire protocol, support for transparent
failover and easy to use blocking API.


* [godoc](https://godoc.org/github.com/optiopay/kafka) generated documentation,
* [code examples](https://godoc.org/github.com/optiopay/kafka#pkg-examples)

## Example

Write all messages from stdin to kafka and print all messages from kafka topic
to stdout.


```go
package main

import (
    "bufio"
    "log"
    "os"
    "strings"

    "github.com/optiopay/kafka"
    "github.com/optiopay/kafka/proto"
)

const (
    topic     = "my-messages"
    partition = 0
)

var kafkaAddrs = []string{"localhost:9092", "localhost:9093"}

// printConsumed read messages from kafka and print them out
func printConsumed(broker kafka.Client) {
    conf := kafka.NewConsumerConf(topic, partition)
    conf.StartOffset = kafka.StartOffsetNewest
    consumer, err := broker.Consumer(conf)
    if err != nil {
        log.Fatalf("cannot create kafka consumer for %s:%d: %s", topic, partition, err)
    }

    for {
        msg, err := consumer.Consume()
        if err != nil {
            if err != kafka.ErrNoData {
                log.Printf("cannot consume %q topic message: %s", topic, err)
            }
            break
        }
        log.Printf("message %d: %s", msg.Offset, msg.Value)
    }
    log.Print("consumer quit")
}

// produceStdin read stdin and send every non empty line as message
func produceStdin(broker kafka.Client) {
    producer := broker.Producer(kafka.NewProducerConf())
    input := bufio.NewReader(os.Stdin)
    for {
        line, err := input.ReadString('\n')
        if err != nil {
            log.Fatalf("input error: %s", err)
        }
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        msg := &proto.Message{Value: []byte(line)}
        if _, err := producer.Produce(topic, partition, msg); err != nil {
            log.Fatalf("cannot produce message to %s:%d: %s", topic, partition, err)
        }
    }
}

func main() {
    // connect to kafka cluster
    broker, err := kafka.Dial(kafkaAddrs, kafka.NewBrokerConf("test-client"))
    if err != nil {
        log.Fatalf("cannot connect to kafka cluster: %s", err)
    }
    defer broker.Close()

    go printConsumed(broker)
    produceStdin(broker)
}
```
