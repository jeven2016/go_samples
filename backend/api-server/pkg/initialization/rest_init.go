package initialization

import (
	"api/app/books"
	common2 "api/pkg/common"
	"api/pkg/middleware"
	"api/pkg/route"
	"github.com/gin-gonic/gin"
)

func SetupEngine(config *common2.Config, app *common2.App) *gin.Engine {
	var engine = gin.Default()
	engine.Use(middleware.GinLogger(app.Log), middleware.GinRecovery(true, app.Log))
	//gin.SetMode(gin.ReleaseMode)
	books.SetupServices(app)
	route.Register(engine, app.Log)
	return engine
}
