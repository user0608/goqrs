package handlers

import (
	"goqrs/envs"
	"goqrs/handlers/collection"
	"goqrs/handlers/events"
	"goqrs/handlers/login"
	"goqrs/repositories"
	"goqrs/security"
	"goqrs/services"
	"goqrs/xstorage"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func StartRoutes(e *echo.Echo) {
	e.GET("/health", health())
	e.GET("/healthy", health())

	loginService := services.NewLoginService(repositories.NewLoginRepository())
	e.POST("/login", login.Handler(loginService))
	e.GET("/events", events.Handler, security.JWTMiddleware)
	e.GET("/events/test", sendEventTest(events.NewEventPublicher()), security.JWTMiddleware)
	collectionRoutes(e.Group("/collection", security.JWTMiddleware))
}
func collectionRoutes(g *echo.Group) {
	baseDir := envs.FindEnv("GOQRS_TEMPLATE_BASE_DIR", "templates")
	documentDir := envs.FindEnv("GOQRS_DOCUMENTS_BASE_DIR", "documents")
	xtemplateStore, err := xstorage.NewTemplateStorage(baseDir)
	if err != nil {
		log.Printf("handlers.collectionRoutes err:%v", err)
		os.Exit(1)
	}
	xdocStore, err := xstorage.NewTemplateStorage(documentDir)
	if err != nil {
		log.Printf("handlers.collectionRoutes err:%v", err)
		os.Exit(1)
	}
	collectionRepo := repositories.NewCollectionRepository()
	service := services.NewCollectionService(collectionRepo, xtemplateStore)
	docService := services.NewDocumentService(
		repositories.NewTicketRepository(),
		collectionRepo, xtemplateStore,
		xdocStore,
	)
	publicher := events.NewEventPublicher()
	g.GET("", collection.HandleFindAll(service))
	g.GET("/:collection_id", collection.HandleFindByID(service))
	g.DELETE("/:collection_id", collection.HandleDeleteCollection(service))

	g.GET("/tags/:collection_id", collection.HandleFindCollectionTags(service))
	g.POST("/tags/:collection_id", collection.HandleAddTag(service))
	g.DELETE("/tag/:tag_id", collection.HandleRemoveTag(service))

	g.POST("/document/test", collection.HandlerPruebaDocument())
	g.GET("/document/generate/:collection_id", collection.HandleProcessDocument(publicher, docService))
	g.GET("/document/:collection_id", collection.DownloadDocumentPDF(xdocStore))
	g.POST("/template/:collection_id", collection.HandleUploadTempleate(service))
	g.GET("/template/:collection_id", collection.ImageTemplate(xtemplateStore))
	g.PUT("", collection.HandelUpdate(service))
	g.POST("", collection.HandelCreate(service))
}
func health() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"message": "success!"})
	}
}
func sendEventTest(p events.EventPublicher) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		p.Publish(username, events.EventData{
			EventName: events.TEST,
			Data:      map[string]string{"message": "success!!!"},
		})
		return nil
	}
}
