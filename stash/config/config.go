package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/service"
)

type (
	Condition struct {
		Key   string
		Value string
		Type  string `json:",default=match,options=match|contains"`
		Op    string `json:",default=and,options=and|or"`
	}

	Filter struct {
		Action     string      `json:",options=drop|remove_field|transfer"`
		Conditions []Condition `json:",optional"`
		Fields     []string    `json:",optional"`
		Field      string      `json:",optional"`
		Target     string      `json:",optional"`
	}

	KafkaConf struct {
		service.ServiceConf
		Brokers    []string
		Group      string
		Topics     []string
		Offset     string `json:",options=first|last,default=last"`
		Conns      int    `json:",default=1"`
		Consumers  int    `json:",default=8"`
		Processors int    `json:",default=8"`
		MinBytes   int    `json:",default=10240"`    // 10K
		MaxBytes   int    `json:",default=10485760"` // 10M
	}

	ClickHouseAuthConf struct {
		Database string
		Username string
		Password string `json:",optional"`
	}

	ClickHouseConf struct {
		Addr          []string
		Auth          ClickHouseAuthConf
		Table         string
		Columns       []string `json:",optional"`
		Interval      int64    `json:"interval,default=15"`
		MaxChunkBytes int      `json:",default=15728640"` // default 15M
	}

	Cluster struct {
		Input struct {
			Kafka KafkaConf
		}
		Filters []Filter `json:",optional"`
		Output  struct {
			Clickhouse ClickHouseConf
		}
	}

	Config struct {
		Clusters    []Cluster
		GracePeriod time.Duration `json:",default=15s"`
	}
)
