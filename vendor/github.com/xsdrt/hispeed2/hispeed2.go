package hispeed2

const version = "1.0.0"

type HiSpeed2 struct {
	AppName string
	Debug   bool
	Version string
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
