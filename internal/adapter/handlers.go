package adapter

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) PutToPost(
	c echo.Context,
) error {
	log.Println("ooooh I was fired babyyy!")
	filepath := c.Param("*")

	if filepath == "" {
		return c.JSON(http.StatusBadRequest, "missing file path")
	}
	filestream := c.Request().Body
	defer filestream.Close()

	parts := strings.SplitN(filepath, "/", 2)

	if len(parts) < 2 {
		return c.JSON(http.StatusBadRequest, "invalid path format, missing jobID")
	}

	jobID := parts[0]
	actualPath := parts[1]
	err := h.svc.MakePostRequest(c.Request().Context(), jobID, actualPath, filestream)
	if err != nil {
		log.Println(err)
	}
	return err
}
