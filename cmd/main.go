package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sirkartik/adapter_sidecar/internal/config"
)

func main() {
	e := echo.New()
	
	cfg := config.NewConfig()

	if err := e.Start(fmt.Sprintf("127.0.0.1:%d", cfg.App.Port)); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
