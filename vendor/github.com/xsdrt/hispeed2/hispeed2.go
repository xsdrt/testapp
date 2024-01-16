package hispeed2

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type HiSpeed2 struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
}

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

	// Creatoe loggers...
	infoLog, errorLog := h.startLogers()
	h.InfoLog = infoLog
	h.ErrorLog = errorLog
	h.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	h.Version = version

	return nil
}

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
