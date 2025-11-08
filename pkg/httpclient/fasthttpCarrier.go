package httpclient

import (
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/propagation"
)

type fasthttpCarrier struct {
	header *fasthttp.RequestHeader
}

func NewFasthttpCarrier(header *fasthttp.RequestHeader) propagation.TextMapCarrier {
	return fasthttpCarrier{header}
}

func (c fasthttpCarrier) Get(key string) string {
	return string(c.header.Peek(key))
}

func (c fasthttpCarrier) Set(key, value string) {
	c.header.Set(key, value)
}

func (c fasthttpCarrier) Keys() []string {
	keys := make([]string, 0)
	c.header.VisitAll(func(k, _ []byte) {
		keys = append(keys, string(k))
	})
	return keys
}
