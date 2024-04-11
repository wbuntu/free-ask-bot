/*
Copyright © 2022 wbuntu
*/
package main

import (
	"gitbub.com/wbuntu/free-ask-bot/cmd"
	_ "gitbub.com/wbuntu/free-ask-bot/docs"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// 手动设置maxprocs来禁用默认的日志打印
	maxprocs.Set()
}

// @title       free-ask-bot API
// @version     1.0
// @description free-ask-bot swagger server.
// @BasePath    /api/v1.0
// @Accept      json
// @Produce     json
func main() {
	cmd.Execute()
}
