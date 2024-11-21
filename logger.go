package goin

import (
	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/zlog"
)

var Log interfaces.Logger

func InitLog(opts ...zlog.Option) {
	var err error
	Log, err = zlog.CreateJsonLogger(opts...)
	if err != nil {
		panic(err)
	}
}
