/*
Copyright © 2022 wbuntu
*/
package main

import (
	"gitbub.com/wbuntu/gin-template/cmd"
	_ "gitbub.com/wbuntu/gin-template/docs"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// 手动设置maxprocs来禁用默认的日志打印
	maxprocs.Set()
}

// @title       gin-template API
// @version     1.0
// @description gin-template swagger server.
// @BasePath    /api/v1.0
// @Accept      json
// @Produce     json
func main() {
	cmd.Execute()
}
