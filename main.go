package main

import (
	"fmt"
	"github.com/ropon/logger"
	"net/http"
	"testgo/conf"
)

// @title work_api
// @version 1.0
// @description 后端快速Api脚手架

// @contact.name Ropon
// @contact.url https://www.ropon.top
// @contact.email ropon@xxx.com

// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host work-api.xxx.com:2345
// @BasePath /
func main() {
	err := conf.Init()
	if err != nil {
		fmt.Printf("init failed, err: %v\n", err)
		return
	}

	logger.Info("Starting...")
	logger.Error("Error log demo...")

	http.HandleFunc("/", sayHello)
	logger.Fatal("Server Start Fatal, Error: %s", http.ListenAndServe(conf.Cfg.Listen, nil))
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello CodoonOps！again")
}
