package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/quexer/sessions"
	"gopkg.in/mgo.v2"
	"net/http"
)

const (
	SESSION_NAME = "web-session"
	SESSION_SALT = "!K.Z[IPnqOXx"
	ADMIN        = "admin"
)

var (
	g_session_store = sessions.NewCookieStore(86400*30*12*10, []byte(SESSION_SALT))
)

func mount(session *mgo.Session, war string) {

	m := martini.Classic()
	m.Handlers(martini.Recovery())

	m.Use(gzip.All())
	m.Use(martini.Static(war, martini.StaticOptions{SkipLogging: true}))
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html", ".shtml"},
	}))

	m.Use(midTextDefault)

	//map web
	m.Use(func(w http.ResponseWriter, c martini.Context) {
		web := &Web{w: w}
		c.Map(web)
	})

	//map ds
	m.Use(func(c martini.Context) {
		session.Refresh()
		ds := &Ds{se: session.Copy()}
		defer ds.se.Close()
		c.Map(ds)
		c.Next()
	})

	m.Use(sessions.Sessions(SESSION_NAME, g_session_store))

	m.Group("/admin/pub", func(r martini.Router) {
		r.Post("/login", binding.Bind(AdminLoginForm{}), adminLoginHandler)
		r.Post("/logout", adminLogoutHandler)
	})

	m.Group("/admin", func(admin martini.Router) {
		admin.Get("/me", adminMeHandler)
		admin.Post("/me/passwd", binding.Bind(UpdatePasswdForm{}), adminUpdatePasswdHandler)
		admin.Group("/op", func(r martini.Router) {
			r.Post("", binding.Bind(OpForm{}), createOpHandler)
			r.Get("", listOpHandler)
			r.Post("/check/login/name", binding.Bind(NameForm{}), checkOpLoginNameHandler)
			r.Post("/reset_password", binding.Bind(ResetOpPasswordForm{}), resetOpPasswordHandler)

			r.Group("/(?P<id>[0-9a-z]{24})", func(r martini.Router) {
				r.Get("", showOpHandler)
				r.Post("", binding.Bind(UpdateVoForm{}), updateOpHandler) //todo 这个貌似就没用。。。
				r.Delete("", delOpHandler)
			}, midOp)
		})
	}, midAdminMe)

	m.Get("/", func() (int, string) {
		return 200, "this is crm"
	})

	http.Handle("/", m)
}
