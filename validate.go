package goin

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/validator/v10"

	ut "github.com/go-playground/universal-translator"
	enTrans "github.com/go-playground/validator/v10/translations/en"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

const (
	Zh = "zh"
	En = "en"
)

type Language string

func InitValidator(language Language) {
	validate = validator.New()

	// 注册自定义校验函数

	// 设置语言环境
	switch language {
	case Zh:
		zhTranslator := zh.New()
		uni := ut.New(zhTranslator, zhTranslator)
		trans, _ = uni.GetTranslator(Zh)
		zhTrans.RegisterDefaultTranslations(validate, trans)
	default:
		enTranslator := en.New()
		uni := ut.New(enTranslator, enTranslator)
		trans, _ = uni.GetTranslator(En)
		enTrans.RegisterDefaultTranslations(validate, trans)
	}

	// 注册自定义标签
	registerCustomTags(validate)
}

func registerCustomTags(validate *validator.Validate) {
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("msg"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}
