package goin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zR-Zr/goin/interfaces"
)

type Context struct {
	*gin.Context
	Logger    interfaces.Logger
	requestID string
}

func NewContext(c *gin.Context, logger interfaces.Logger) *Context {
	// 生成请求ID
	requestID := uuid.New().String()

	// 初始化链路追中器

	return &Context{
		Context:   c,
		Logger:    logger,
		requestID: requestID,
	}
}

// QueryString String 获取请求参数 ?username=xxx
// key: 参数名
func (c *Context) QueryString(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryInt Int 获取请求参数 ?userID=1020301230
func (c *Context) QueryInt(key string) (int, error) {
	value := c.QueryString(key)
	return strconv.Atoi(value)
}

// QueryBool Bool 获取请求参数
func (c *Context) QueryBool(key string) (bool, error) {
	value := c.QueryString(key)
	return strconv.ParseBool(value)
}

func (c *Context) PathParam(key string) string {
	value := c.Param(key)
	return value
}

func (c *Context) PathParamInt(key string, defaultValue ...int) int {
	value, err := strconv.Atoi(c.Param(key))
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// RequestID 获取 请求 ID, 每个请求都是唯一的
func (c *Context) RequestID() string {
	return c.requestID
}

func (c *Context) AbortWithError(e error) {
	c.Context.Error(e)
}

func (c *Context) Success(msg string, payload any) {
	resp := map[string]any{
		"code":    2000,
		"message": msg,
	}

	if payload != nil {
		resp["data"] = payload
	}

	c.JSON(http.StatusOK, resp)
}

// Vaidate 校验请求参数
// obj: 需要被校验的对象
func (c *Context) Validate(obj any) error {
	// 执行校验
	if err := validate.Struct(obj); err != nil {
		errs := err.(validator.ValidationErrors)
		var sliceErrs []string
		for _, e := range errs {
			// 使用自定义标签获取字段名
			fieldName := e.Field()
			// jsonTagName: 获取该字段的 json , uri, form 的名称
			tag := getTagName(obj, e.StructField())

			msg := extractValidationMessage(fieldName, e)

			// 格式化错误信息
			errMsg := fmt.Sprintf("%s:%s", tag, msg)
			sliceErrs = append(sliceErrs, errMsg)
		}

		return fmt.Errorf(strings.Join(sliceErrs, ","))
	}
	return nil
}

func extractValidationMessage(fieldName string, e validator.FieldError) string {
	errorsMap := parseErrorsToMap(fieldName)

	// 如果有自定义的错误消息,则使用自定义消息
	if msg, exists := errorsMap[e.Tag()]; exists && msg != "" {
		return msg
	}

	// 如果没有自定义的错误消息,则使用 validator 的默认错误小学
	msg := e.Translate(trans)

	// 去除字段名
	return strings.TrimPrefix(msg, fieldName)
}

func parseErrorsToMap(str string) map[string]string {
	errorsMap := map[string]string{}

	kvStringSlice := strings.Split(str, ";")
	for _, kvStr := range kvStringSlice {
		kv := strings.Split(kvStr, ":")
		if len(kv) == 2 {
			errorsMap[kv[0]] = kv[1]
		}
	}
	return errorsMap
}

func (c *Context) defaultLogFields() map[string]any {
	return map[string]any{
		"request_id": c.requestID,
		"url":        c.Context.Request.URL.String(),
		"method":     c.Context.Request.Method,
		"ip":         c.Context.ClientIP(),
	}
}

func convertMapToSlice(fields map[string]any) []any {
	result := make([]any, 0, len(fields)*2)

	for k, v := range fields {
		result = append(result, k, v)
	}
	return result
}
