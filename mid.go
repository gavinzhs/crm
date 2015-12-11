package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/quexer/sessions"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
)

const (
	CONTENT_TYPE = "Content-Type"
)

func midTextDefault(w http.ResponseWriter) {
	if w.Header().Get(CONTENT_TYPE) == "" {
		w.Header().Set(CONTENT_TYPE, "text/plain; charset=UTF-8")
	}
}

type OpSession struct {
	me *Operator
}

func midAdminMe(c martini.Context, rd render.Render, session sessions.Session, ds *Ds) {
	name := session.Get(ADMIN)
	if name == nil {
		rd.Error(401)
		return
	}

	var me *Operator
	err := ds.se.DB(DB).C(C_OPERATOR).Find(bson.M{"login": name.(string)}).One(&me)
	if err != nil {
		rd.Error(401)
		return
	}

	opSession := &OpSession{me}
	c.Map(opSession)

}

func midOp(param martini.Params, rd render.Render, c martini.Context, ds *Ds) {
	id := param["id"]
	o, err := loadOperator(ds, bson.ObjectIdHex(id))
	if notFound(err) {
		rd.Error(404)
		return
	}

	chk(err)
	c.Map(o)
}

func midOrg(param martini.Params, w http.ResponseWriter, c martini.Context, ds *Ds) {
	oid := param["oid"]
	id, _ := strconv.Atoi(oid)

	var org *Org
	var err error
	if id == 0 {
		org, err = loadOrgByQuery(ds.se, bson.M{"parent": id})
	} else {
		org, err = loadOrg(ds.se, id)
	}

	if err == mgo.ErrNotFound {
		http.Error(w, "org not found", 404)
		return
	}
	chk(err)
	c.Map(org)
}
