package session

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

// using the alexedwards/scs package for session control...
type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieName     string
	CookieDomain   string
	SessionType    string
	CookieSecure   string
	DBPool         *sql.DB
}

func (h *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should session last?
	minutes, err := strconv.Atoi(h.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// should cookies consist?
	if strings.ToLower(h.CookiePersist) == "true" {
		persist = true
	}

	// must cookies be secure?
	if strings.ToLower(h.CookieSecure) == "true" {
		secure = true
	}

	// create session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = h.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = h.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?
	switch strings.ToLower(h.SessionType) {
	case "redis":

	case "mysql", "mariadb":
		session.Store = mysqlstore.New(h.DBPool)

	case "postgres", "postgresql":
		session.Store = postgresstore.New(h.DBPool)

	default:
		//cookie
	}

	return session

}
