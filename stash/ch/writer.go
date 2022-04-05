package ch

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ilyq69/go-stash/stash/config"
	"github.com/zeromicro/go-zero/core/executors"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	Writer struct {
		client      clickhouse.Conn
		ctx         context.Context
		columns     []string
		columnsType map[string]string
		query       string
		database    string
		table       string
		inserter    *executors.ChunkExecutor
	}

	ValueWithIndex struct {
		val []interface{}
	}

	ClickHouseDescType struct {
		Name              string `db:"name"`
		Type              string `db:"type"`
		DefaultType       string `db:"default_type"`
		DefaultExpression string `db:"default_expression"`
		Comment           string `db:"comment"`
		CodecExpression   string `db:"codec_expression"`
		TTLExpression     string `db:"ttl_expression"`
	}
)

func NewWriter(c config.ClickHouseConf) (*Writer, error) {
	client, err := clickhouse.Open(&clickhouse.Options{
		Addr: c.Addr,
		Auth: clickhouse.Auth{
			Database: c.Auth.Database,
			Username: c.Auth.Username,
			Password: c.Auth.Password,
		},
	})
	if err != nil {
		return nil, err
	}

	var query string
	if len(c.Columns) > 0 {
		query = "INSERT INTO " + c.Auth.Database + "." + c.Table + " (" + strings.Join(c.Columns, ", ") + ")"
	} else {
		query = "INSERT INTO " + c.Auth.Database + "." + c.Table
	}

	writer := Writer{
		client:   client,
		ctx:      context.Background(),
		database: c.Auth.Database,
		table:    c.Table,
		query:    query,
		columns:  c.Columns,
	}
	writer.inserter = executors.NewChunkExecutor(writer.execute, executors.WithChunkBytes(c.MaxChunkBytes), executors.WithFlushInterval(time.Duration(c.Interval)*time.Second))

	if err = writer.clickhouseColumns(); err != nil {
		return nil, err
	}

	return &writer, nil
}

func (w *Writer) clickhouseTableDesc() ([]*ClickHouseDescType, error) {
	var descTypes []*ClickHouseDescType
	rows, err := w.client.Query(w.ctx, "DESC "+w.database+"."+w.table)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var descType ClickHouseDescType
		err = rows.Scan(&descType.Name, &descType.Type, &descType.DefaultType, &descType.DefaultExpression, &descType.Comment, &descType.CodecExpression, &descType.TTLExpression)
		if err != nil {
			return nil, err
		}

		descTypes = append(descTypes, &descType)
	}
	defer rows.Close()

	return descTypes, nil
}

func (w *Writer) clickhouseColumns() error {
	descTypes, err := w.clickhouseTableDesc()
	if err != nil {
		return err
	}

	columnType := make(map[string]string)
	for _, item := range descTypes {
		columnType[item.Name] = item.Type
	}
	w.columnsType = columnType

	if len(w.columns) == 0 {
		for _, desc := range descTypes {
			w.columns = append(w.columns, desc.Name)
		}
	}

	for _, column := range w.columns {
		_, ok := columnType[column]
		if !ok {
			return errors.New(column + " not in " + w.database + " " + w.table)
		}
	}
	return nil
}

func (w *Writer) Write(val map[string]interface{}) error {
	v, err := w.PrepareData(val)
	if err != nil {
		logx.Alert(err.Error())
		return nil
	}
	return w.inserter.Add(ValueWithIndex{
		val: v,
	}, len(val))
}

func (w *Writer) PrepareData(val map[string]interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(w.columns))
	for index, column := range w.columns {
		c, ok := val[column]
		if !ok {
			return nil, errors.New(column + " not in data")
		}
		v, err := toClickhouseType(c, w.columnsType[column])
		if err != nil {
			return nil, err
		}
		result[index] = v
	}

	return result, nil
}

func (w *Writer) execute(vals []interface{}) {

	bulk, err := w.client.PrepareBatch(w.ctx, w.query)
	if err != nil {
		logx.Error(err)
		return
	}
	for _, val := range vals {
		err = bulk.Append(val.(ValueWithIndex).val...)
		if err != nil {
			logx.Error(err)
			return
		}
	}

	err = bulk.Send()
	if err != nil {
		logx.Error(err)
	}
	logx.Infof("insert %d rows", len(vals))
}
