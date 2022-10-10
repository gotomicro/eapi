package bffcore

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/etrace"

	"github.com/gin-gonic/gin"
)

// HandlerFunc core封装后的handler
type HandlerFunc func(c *Context)

var pbMarshaler = jsonpb.Marshaler{
	EmitDefaults: true,
	OrigName:     true,
	// EnumsAsInts:  true,
}

type Pagination struct {
	// 当前页数
	CurrentPage int32 `json:"currentPage" form:"currentPage"`
	// 每页总数
	PageSize int32 `json:"pageSize" form:"pageSize"`
	// 排序字符串
	Sort string `json:"sort" form:"sort"`
}

type ListPage struct {
	// 列表
	List any `json:"list"`
}

// Handle 将core.HandlerFunc转换为gin.HandlerFunc
func Handle(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{
			c,
		}
		h(ctx)
	}
}

// Context core封装后的Context
type Context struct {
	*gin.Context
}

const (
	// CodeOK 表示响应成功状态码
	CodeOK = 0
	// CodeErr 表示默认响应失败状态码
	CodeErr     = 1
	CodeInvalid = 401
)

// Ctx 返回Request.Context()
func (c *Context) Ctx() context.Context {
	return c.Request.Context()
}

// Res 标准JSON输出格式
type Res struct {
	// Code 响应的业务错误码。0表示业务执行成功，非0表示业务执行失败。
	Code int `json:"code"`
	// Msg 响应的参考消息。前端可使用msg来做提示
	Msg string `json:"msg"`
	// Data 响应的具体数据
	Data any `json:"data"`
}

// ResProto 标准JSON输出格式
type ResProto struct {
	// Code 响应的业务错误码。0表示业务执行成功，非0表示业务执行失败。
	Code int `json:"code"`
	// Msg 响应的参考消息。前端可使用msg来做提示
	Msg string `json:"msg"`
	// Data 响应的具体数据
	Data json.RawMessage `json:"data"`
}

// JSON 输出响应JSON
// 形如 {"code":<code>, "msg":<msg>, "data":<data>}
func (c *Context) JSON(httpStatus int, res Res) {
	c.Context.JSON(httpStatus, res)
}

// JSONOK 输出响应成功JSON，如果data不为零值，则输出data
// 形如 {"code":0, "msg":"成功", "data":<data>}
func (c *Context) JSONOK(data ...interface{}) {
	j := new(Res)
	j.Code = CodeOK
	j.Msg = "成功"
	if len(data) > 0 {
		j.Data = data[0]
	} else {
		j.Data = ""
	}
	c.Context.JSON(http.StatusOK, j)
	return
}

// ProtoJSONOK 输出响应成功JSON，如果data不为零值，则输出data
// 形如 {"code":0, "msg":"成功", "data":<data>}
func (c *Context) ProtoJSONOK(data proto.Message) {
	j := new(ResProto)
	j.Code = CodeOK
	j.Msg = "成功"
	_buffer := new(bytes.Buffer)
	_ = pbMarshaler.Marshal(_buffer, data)
	j.Data = json.RawMessage(_buffer.String())
	c.Context.JSON(http.StatusOK, j)
	return
}

// JSONE 输出失败响应
// 形如 {"code":<code>, "msg":<msg>, "data":<data>}
func (c *Context) JSONE(code int, msg string, data interface{}) {
	j := new(Res)
	j.Code = code
	j.Msg = msg

	if econf.GetBool("debug") == true {
		switch d := data.(type) {
		case error:
			j.Data = d.Error()
		default:
			j.Data = data
		}
	}

	elog.Warn("biz warning", elog.FieldValue(msg), elog.FieldValueAny(data), elog.FieldTid(etrace.ExtractTraceID(c.Request.Context())))
	c.Context.JSON(http.StatusOK, j)
	return
}

// Bind 将请求消息绑定到指定对象中，并做数据校验。如果校验失败，则返回校验失败错误中文文案
// 并将HTTP状态码设置成400
func (c *Context) Bind(obj interface{}) (err error) {
	return validate(c.Context.Bind(obj))
}

// ShouldBind 将请求消息绑定到指定对象中，并做数据校验。如果校验失败，则返回校验失败错误中文文案
// 类似Bind，但是不会将HTTP状态码设置成400
func (c *Context) ShouldBind(obj interface{}) (err error) {
	return validate(c.Context.ShouldBind(obj))
}
