package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	api "stock_assistant/backend/gateway/biz/handler/api"
)

// RegisterManualRoutes registers manually added routes
func RegisterManualRoutes(r *server.Hertz) {
	apiGroup := r.Group("/api")
	apiGroup.GET("/graph/neighborhood", api.GetGraphNeighborhood)
	apiGroup.GET("/graph/entity", api.GetGraphEntityProfile)
	apiGroup.GET("/graph/events", api.SearchGraphEvents)
}
