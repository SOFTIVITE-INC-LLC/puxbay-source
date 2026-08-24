package admin

import (
	"github.com/gin-gonic/gin"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/themes/adminlte"

	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
	"github.com/GoAdminGroup/go-admin/plugins/admin"

	appconfig "github.com/softivite/puxbay/internal/config"
)

// Setup initializes GoAdmin.
func Setup(r *gin.Engine, appCfg *appconfig.Config) error {
	eng := engine.Default()

	cfg := config.Config{
		Databases: config.DatabaseList{
			"default": {
				Host:   appCfg.Database.Host,
				Port:   appCfg.Database.Port,
				User:   appCfg.Database.User,
				Pwd:    appCfg.Database.Password,
				Name:   appCfg.Database.DBName,
				Driver: config.DriverPostgresql,
				Params: map[string]string{
					"sslmode": appCfg.Database.SSLMode,
				},
			},
		},
		UrlPrefix: "admin",
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Language:    language.EN,
		IndexUrl:    "/",
		Debug:       appCfg.App.Env == "development",
		ColorScheme: adminlte.ColorschemeSkinBlack,
		Theme:       "adminlte",
	}

	adminPlugin := admin.NewAdmin(CreateGenerators())

	// Add GoAdmin template

	// We use the Gin adapter
	if err := eng.AddConfig(&cfg).
		AddPlugins(adminPlugin).
		Use(r); err != nil {
		return err
	}

	return nil
}
