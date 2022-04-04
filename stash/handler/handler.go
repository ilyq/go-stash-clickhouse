package handler

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kevwan/go-stash/stash/ch"
	"github.com/kevwan/go-stash/stash/filter"
)

type MessageHandler struct {
	writer  *ch.Writer
	filters []filter.FilterFunc
}

func NewHandler(writer *ch.Writer) *MessageHandler {
	return &MessageHandler{
		writer: writer,
	}
}

func (mh *MessageHandler) AddFilters(filters ...filter.FilterFunc) {
	for _, f := range filters {
		mh.filters = append(mh.filters, f)
	}
}

func (mh *MessageHandler) Consume(_, val string) error {
	var m map[string]interface{}
	if err := jsoniter.Unmarshal([]byte(val), &m); err != nil {
		return err
	}
	for _, proc := range mh.filters {
		if m = proc(m); m == nil {
			return nil
		}
	}

	return mh.writer.Write(m)
}
