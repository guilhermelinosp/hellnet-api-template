package hello

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guilhermelinosp/fast-platform/platform"
)

// maxBodyBytes bounds the JSON body of POST /hello.
const maxBodyBytes = 1 << 20

// Handler exposes the greeting endpoints on a gin router.
type Handler struct {
	service Service
}

// NewHandler wires the handler to its service.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register mounts this module's routes (usually on the /api/v1 group).
// Three flavors on purpose — they document every input style a typical
// endpoint needs: query string, path parameter and JSON body.
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/hello", h.greetByQuery)
	r.GET("/hello/:name", h.greetByPath)
	r.POST("/hello", h.greetByBody)
}

// greetResponse is the wire shape of a successful greeting.
type greetResponse struct {
	Message string `json:"message"`
}

func (h *Handler) greetByQuery(c *gin.Context) {
	h.respond(c, http.StatusOK, normalize(c.Query("name")))
}

func (h *Handler) greetByPath(c *gin.Context) {
	name := normalize(c.Param("name"))
	if name == "" {
		platform.AbortError(c, platform.ValidationError("name", "path parameter is required"))
		return
	}
	h.respond(c, http.StatusOK, name)
}

// greetRequest is the strict body contract for POST /hello: unknown fields are
// rejected (typo-proof API by default).
type greetRequest struct {
	Name string `json:"name"`
}

func (h *Handler) greetByBody(c *gin.Context) {
	var in greetRequest
	dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		platform.AbortError(c, platform.ValidationError("body", "must be a valid JSON object with only the documented fields"))
		return
	}
	name := normalize(in.Name)
	if name == "" {
		platform.AbortError(c, platform.ValidationError("name", "is required"))
		return
	}
	h.respond(c, http.StatusCreated, name)
}

func (h *Handler) respond(c *gin.Context, status int, name string) {
	msg, err := h.service.Greet(c.Request.Context(), name)
	if err != nil {
		platform.AbortError(c, err)
		return
	}
	c.JSON(status, greetResponse{Message: msg})
}

func normalize(name string) string { return strings.TrimSpace(name) }
