package model

import "strings"

// Code 响应码
type Code string

// TODO：从注释自动生成Message方法
func (c Code) Message() string {
	var msg string
	switch c {
	case CodeSuccess:
		msg = "调用成功"
	case CodeInternalError:
		msg = "内部错误"
	case CodeParamError:
		msg = "参数错误"
	case CodeNotExists:
		msg = "资源不存在"
	case CodeAlreadyExists:
		msg = "资源已存在"
	case CodeForbidOperate:
		msg = "禁止操作"
	default:
		msg = "Unknown"
	}
	return msg
}

func response(code Code, messages ...string) BaseResponse {
	r := BaseResponse{Code: code, Message: code.Message()}
	if len(messages) > 0 {
		r.Message += ": " + strings.Join(messages, ": ")
	}
	return r
}

const (
	CodeSuccess       = Code("Success")       // 调用成功
	CodeInternalError = Code("InternalError") // 内部错误
	CodeParamError    = Code("ParamError")    // 参数错误
	CodeNotExists     = Code("NotExists")     // 资源不存在
	CodeAlreadyExists = Code("AlreadyExists") // 资源已存在
	CodeForbidOperate = Code("ForbidOperate") // 禁止操作
)

var (
	RespSuccess = response(CodeSuccess)
)
