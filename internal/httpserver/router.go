package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/teaspeak-v2/wt-bot-ms-bots-v1/docs"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/handlers"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/httpserver/middleware"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/token"
)

type RouterDeps struct {
	Bots           *handlers.BotHandler
	Health         *handlers.HealthHandler
	Tokens         *token.Manager
	ServiceAPIKey  string
	AllowedOrigins []string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recover(nil), middleware.Logger(nil), middleware.CORS(deps.AllowedOrigins))
	r.Get("/", deps.Health.Root)
	r.Get("/healthz", deps.Health.Healthz)
	r.Get("/readyz", deps.Health.Readyz)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(auth chi.Router) {
			auth.Use(middleware.Auth(deps.Tokens))

			auth.Get("/bots", deps.Bots.List)
			auth.Post("/bots", deps.Bots.Create)
			auth.Get("/bots/{id}", deps.Bots.GetByID)
			auth.Patch("/bots/{id}", deps.Bots.Update)
			auth.Delete("/bots/{id}", deps.Bots.Delete)
			auth.Post("/bots/{id}/start", deps.Bots.Start)
			auth.Post("/bots/{id}/stop", deps.Bots.Stop)
		})

		api.Group(func(svc chi.Router) {
			svc.Use(middleware.ServiceKey(deps.ServiceAPIKey))

			svc.Get("/bots/{id}/config", deps.Bots.Config)
			svc.Post("/bots/{id}/status", deps.Bots.UpdateStatus)
		})

		api.Group(func(admin chi.Router) {
			admin.Use(middleware.Auth(deps.Tokens), middleware.RequireRole(models.RoleAdmin))

			_ = admin
		})
	})

	return r
}
