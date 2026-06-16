package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (api *API) Router() http.Handler {
	router := chi.NewRouter()
	router.Use(api.requestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Get("/health", api.health)
	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", api.login)

		r.Group(func(r chi.Router) {
			r.Use(api.auth)
			r.Post("/auth/logout", api.logout)
			r.Get("/auth/me", api.me)

			r.Get("/knowledge-bases", api.listKnowledgeBases)
			r.With(api.requirePermission("knowledge:write")).Post("/knowledge-bases", api.createKnowledgeBase)
			r.With(api.requirePermission("knowledge:write")).Patch("/knowledge-bases/{id}", api.updateKnowledgeBase)
			r.Get("/knowledge-bases/{id}/documents", api.listDocuments)
			r.With(api.requirePermission("knowledge:write")).Post("/knowledge-bases/{id}/documents", api.uploadDocument)

			r.Get("/ingestion-jobs/{id}", api.getIngestionJob)
			r.With(api.requirePermission("knowledge:write")).Post("/ingestion-jobs/{id}:retry", api.retryIngestionJob)

			r.Get("/conversations", api.listConversations)
			r.Post("/conversations", api.createConversation)
			r.Get("/conversations/{id}/messages", api.listMessages)
			r.Post("/conversations/{id}/messages:stream", api.streamMessage)
			r.Get("/messages/{id}/details", api.messageDetails)
		})
	})

	return router
}
