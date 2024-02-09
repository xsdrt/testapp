package hispeed2

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/xsdrt/hispeed2/render"
)

const version = "1.0.0"

// Hispeed2  is the overall type for the Hispeed2 package... Exported members in tghis type
// are available to any applicatiomn that uses it...Other than the config as no reason for anybody using Hispeed 2 needs to know this info...
type HiSpeed2 struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	JetViews *jet.Set
	config   config
}

type config struct {
	port     string
	renderer string // What template engine to use , either the std Go or Jet pkg...
}

// New reads the .env file, creates our app config, populates the HiSpeed2 type with settings
// based on the .env values, and creates necessary folders and files if they don't exist...
func (h *HiSpeed2) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := h.Init(pathConfig)
	if err != nil {
		return err
	}

	err = h.checkDotEnv(rootPath) // Check the root path of the application (or TestApp during development)...
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// Create loggers...
	infoLog, errorLog := h.startLogers()
	h.InfoLog = infoLog
	h.ErrorLog = errorLog
	h.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	h.Version = version
	h.RootPath = rootPath
	h.Routes = h.routes().(*chi.Mux) // Cast to a pointer of chi.Mux...

	h.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}
	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	h.JetViews = views

	h.createRenderer()

	return nil
}

// Init creates necessary folders for our HiSpped2 application...
func (h *HiSpeed2) Init(p initPaths) error {
	root := p.rootPath //holds the full root path to the web app...
	for _, path := range p.folderNames {
		// create the folder if it doesn't exist...
		err := h.CreateDirIfNotExist(root + "/" + path) // creates the dir if not exists...
		if err != nil {
			return err
		}
	}
	return nil
}

// ListenAndServe starts the web server...
func (h *HiSpeed2) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     h.ErrorLog,
		Handler:      h.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second, // Longtime out for dev purposes for now...
	}

	h.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	h.ErrorLog.Fatal(err)
}

func (h *HiSpeed2) checkDotEnv(path string) error {
	err := h.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path)) // look into the root lvl of app to see if the env file exist, if not return an err...
	if err != nil {
		return err
	}
	return nil
}

func (h *HiSpeed2) startLogers() (*log.Logger, *log.Logger) { // made some vars, when in prod and not debug will write to files
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (h *HiSpeed2) createRenderer() {
	myRenderer := render.Render{
		Renderer: h.config.renderer,
		RootPath: h.RootPath,
		Port:     h.config.port,
		JetViews: h.JetViews,
	}
	h.Render = &myRenderer

}
