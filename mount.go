package main
import (
    "gopkg.in/mgo.v2"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "github.com/martini-contrib/gzip"
    "net/http"
    "github.com/quexer/sessions")

const (
    SESSION_NAME       = "web-session"
    SESSION_SALT       = "!K.Z[IPnqOXx"
)

var (
    g_session_store = sessions.NewCookieStore(86400*30*12*10, []byte(SESSION_SALT))
)

func mount(session mgo.Session, war string){

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

    m.Get("/", func() (int, string) {
        return 200, "this is crm"
    })
}
