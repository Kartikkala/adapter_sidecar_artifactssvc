package adapter

import "github.com/labstack/echo/v4"

func AttachRoutes(e *echo.Echo, svc *Service) {
	h := NewHandler(svc)
	e.PUT("/hls/*", h.PutToPost)
}
