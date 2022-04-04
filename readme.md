English | [简体中文](readme-cn.md)

# go-stash

go-stash is a high performance, free and open source server-side data processing pipeline that ingests data from Kafka, processes it, and then sends it to Clickhouse. 

![go-stash](doc/flow.png)

## Quick Start

### Install

```shell
cd stash && go build stash.go
```

### Quick Start

- With binary

```shell
./stash -f etc/config.yaml
```

The config.yaml example is as follows:

```yaml
Clusters:
- Input:
    Kafka:
      Name: go-stash
      Log:
        Mode: file
      Brokers:
      - "172.16.48.41:9092"
      - "172.16.48.42:9092"
      - "172.16.48.43:9092"
      Topic: ngapplog
      Group: stash
      Conns: 3
      Consumers: 10
      Processors: 60
      MinBytes: 1048576
      MaxBytes: 10485760
      Offset: first
  Filters:
  - Action: drop
    Conditions:
      - Key: status
        Value: 503
        Type: contains
      - Key: type
        Value: "app"
        Type: match
        Op: and
  - Action: remove_field
    Fields:
    - message
    - source
    - beat
    - fields
    - input_type
    - offset
    - "@version"
    - _score
    - _type
    - clientip
    - http_host
    - request_time
  Output:
    Clickhouse:
      Addr:
        - "127.0.0.1:9000"
      Auth:
        Database: default
        Username: default
        Password:
      Table: example
      Columns:
        - Col1
        - Col2
        - Col3
```

## Details

### input

```yaml
Conns: 3
Consumers: 10
Processors: 60
MinBytes: 1048576
MaxBytes: 10485760
Offset: first
```
#### Conns
* The number of links to kafka, the number of links is based on the number of cores of the CPU, usually <= the number of cores of the CPU.

#### Consumers
* The number of open threads per connection, the calculation rule is Conns * Consumers, not recommended to exceed the total number of slices, for example, if the topic slice is 30, Conns * Consumers <= 30

#### Processors
* The number of threads to process data, depending on the number of CPU cores, can be increased appropriately, the recommended configuration: Conns * Consumers * 2 or Conns * Consumers * 3, for example: 60 or 90

#### MinBytes MaxBytes
* The default size of the data block from kafka is 1M~10M. If the network and IO are better, you can adjust it higher.

#### Offset
* Optional last and false, the default is last, which means read data from kafka from the beginning


### Filters

```yaml
- Action: drop
  Conditions:
    - Key: k8s_container_name
      Value: "-rpc"
      Type: contains
    - Key: level
      Value: info
      Type: match
      Op: and
- Action: remove_field
  Fields:
    - message
    - _source
    - _type
    - _score
    - _id
    - "@version"
    - topic
    - index
    - beat
    - docker_container
    - offset
    - prospector
    - source
    - stream
- Action: transfer
  Field: message
  Target: data
```

#### - Action: drop
  - Delete flag: The data that meets this condition will be removed when processing and will not be entered into es
  - According to the delete condition, specify the value of the key field and Value, the Type field can be contains (contains) or match (match)
  - Splice condition Op: and, can also write or

#### - Action: remove_field
  Remove_field_id: the field to be removed, just list it below

#### - Action: transfer
  Transfer field identifier: for example, the message field can be redefined as a data field


### Output

#### Columns
* Optional parameters

#### MaxChunkBytes
* The size of the bulk submitted each time, default is 15M

#### GracePeriod
* The default is 15s, which is used to process the remaining consumption and data within 15s after the program closes and exits gracefully


### Support types
 - [x] Float32, Float64
 - [x] Int8, Int16, Int32, Int64
 - [x] UInt8, UInt16, UInt32, UInt64
 - [x] IPv4, IPv6
 - [x] Bool, Boolean
 - [x] Date, Date32, DateTime
 - [x] UUID
 - [x] String

### Reference
 - [https://github.com/kevwan/go-stash](https://github.com/kevwan/go-stash)