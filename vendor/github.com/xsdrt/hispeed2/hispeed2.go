package hispeed2

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/xsdrt/hispeed2/cache"
	"github.com/xsdrt/hispeed2/render"
	"github.com/xsdrt/hispeed2/session"
)

const version = "1.0.0"

var useRedisCache *cache.RedisCache

// Hispeed2  is the overall type for the Hispeed2 package... Exported members in tghis type
// are available to any applicatiomn that uses it...Other than the config as no reason for anybody using Hispeed 2 needs to know this info...
type HiSpeed2 struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
}

type config struct {
	port        string
	renderer    string // What template engine to use , either the std Go or Jet pkg...
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
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

	// connect to the database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := h.OpenDB(os.Getenv("DATABASE_TYPE"), h.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		h.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" { // Check to see if need to connect to redis; user might not be using redis...
		useRedisCache = h.createClientRedisCache()
		h.Cache = useRedisCache
	}

	h.InfoLog = infoLog
	h.ErrorLog = errorLog
	h.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	h.Version = version
	h.RootPath = rootPath
	h.Routes = h.routes().(*chi.Mux) // Cast to a pointer of chi.Mux...

	h.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      h.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// create a session...

	sess := session.Session{
		CookieLifetime: h.config.cookie.lifetime,
		CookiePersist:  h.config.cookie.persist,
		CookieName:     h.config.cookie.name,
		SessionType:    h.config.sessionType,
		CookieDomain:   h.config.cookie.domain,
		//DBPool:         h.DB.Pool,  // this might be set to nil if not using a db
	}

	switch h.config.sessionType {
	case "redis":
		sess.RedisPool = useRedisCache.Conn

	case "mysql", "mariadb", "postgres", "postgresql":
		sess.DBPool = h.DB.Pool

	}

	h.Session = sess.InitSession()
	h.EncryptionKey = os.Getenv("KEY")

	if h.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(), // allows to work on jet template with out restarting app after changes
		)

		h.JetViews = views

	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)

		h.JetViews = views

	}

	h.createRenderer()

	return nil
}

// Init creates necessary folders for our HiSpeed2 application...
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

	defer h.DB.Pool.Close()

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
		Session:  h.Session,
	}
	h.Render = &myRenderer

}

// Create a Redis cacheClient
func (h *HiSpeed2) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   h.createRedisPool(),
		Prefix: h.config.redis.prefix,
	}
	return &cacheClient
}

// Create a Redis Pool
func (h *HiSpeed2) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				h.config.redis.host,
				redis.DialPassword(h.config.redis.password))
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (h *HiSpeed2) BuildDSN() string {
	var dsn string //store the connection string in this var...

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" { //need to support some postgres that do not require a password...
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}

	return dsn

}
