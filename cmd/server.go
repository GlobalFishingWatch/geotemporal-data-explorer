package cmd

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/4wings/cli/assets"
	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/database"
	"github.com/4wings/cli/internal/middlewares"
	"github.com/4wings/cli/internal/routes"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var server = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Long:  `Server`,
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		exit := make(chan os.Signal)
		signal.Notify(exit, os.Interrupt)

		go func() {
			for {
				wg.Add(1)
				go func() {
					RunServer(port, true)

					log.Debug("Restarting server")
					wg.Done()
				}()
				wg.Wait()
			}
		}()
		<-exit
		types.Quit <- os.Kill

	},
}

var port int

func init() {
	server.Flags().IntVarP(&port, "port", "p", 8080, "Port")
	server.Flags().StringP("local-db", "", "./local.db", "Directory of local db")
	server.Flags().StringP("gee-account-file", "", "", "Path to key file of service account to use for Google Earth Engine datasets")
	server.Flags().StringP("gfw-token", "", "", "Token of Global Fishing Watch API")

	viper.BindPFlag("local-db", server.Flags().Lookup("local-db"))
	viper.BindPFlag("gee-account-file", server.Flags().Lookup("gee-account-file"))
	viper.BindPFlag("gfw-token", server.Flags().Lookup("gfw-token"))

	rootCmd.AddCommand(server)
}

func RunServer(port int, local bool) {
	viper.Set("local", true)
	viper.Set("gee", false)
	if viper.GetString("gee-account-file") != "" {
		log.Debugf("Reading gee account file %s", viper.GetString("gee-account-file"))
		geeAccount, err := utils.ReadFile(viper.GetString("gee-account-file"))
		if err != nil {
			log.Errorf("error reading gee account file %e", err)
		} else {
			viper.Set("gee", true)
			viper.Set("gee-account", geeAccount)
		}
	}
	if viper.GetString("gfw-token") != "" {
		viper.Set("gfw", true)
	}

	log.Debug("Loading server")
	r := gin.Default()

	log.Debug("Opening database")
	err := database.Open()
	if err != nil {
		panic(err)
	}

	r.Use(gin.Recovery())
	r.Use(middlewares.CorsMiddleware)
	r.Use(middlewares.ErrorHandle())

	if local {

		templ := template.Must(template.New("").ParseFS(assets.F, "*.html"))
		r.SetHTMLTemplate(templ)
		r.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "Main website",
			})
		})
		r.GET("/v1/close", func(c *gin.Context) {
			database.LocalDB.Close()
			c.JSON(http.StatusOK, gin.H{"ok": 1})
		})
		r.StaticFS("/data-explorer", http.FS(assets.F))
		r.GET("/v1/datasets", routes.GetAllDatasets)
		r.GET("/v1/datasets/:id", routes.GetDataset)
		r.DELETE("/v1/datasets/:id", routes.DeleteDataset)
		r.POST("/v1/datasets", routes.CreateDataset)
		r.GET("/v1/datasets/:id/data", routes.GetContextData)
		r.GET("/v1/datasets/:id/filters", routes.GetFiltersData)
		r.POST("/v1/files", routes.UploadFile)
		r.GET("/v1/files/:filename/status", routes.GetTempFile)
		r.GET("/v1/files/:filename/fields", routes.GettingFieldsFile)

	}
	r.Use(middlewares.DatasetMiddleware)
	r.GET("/v1/4wings/tile/heatmap/:z/:x/:y", middlewares.CheckQueryParams(internal.TILE_QUERY_PARAMS_V1), middlewares.IsValidZoomInDatasetsMiddleware, routes.Tile)

	listenAddr := fmt.Sprintf(":%d", port)
	log.Info("Listening in ", listenAddr)

	s := &http.Server{
		Addr:           listenAddr,
		Handler:        r,
		ReadTimeout:    time.Duration(300) * time.Second,
		WriteTimeout:   time.Duration(300) * time.Second,
		MaxHeaderBytes: 1 << 60,
	}

	go func() {
		// service connections
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.

	signal.Notify(types.Quit, os.Interrupt)
	<-types.Quit
	log.Debug("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")
}
