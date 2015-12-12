package main

import (
	"github.com/go-martini/martini"
	"github.com/quexer/sessions"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strings"
)

func adminLoginHandler(r *http.Request, web *Web, ds *Ds, session sessions.Session) (int, string) {

	name := strings.TrimSpace(r.PostFormValue("name"))
	passwd := strings.TrimSpace(r.PostFormValue("passwd"))

	if name == "" || passwd == "" {
		return 400, "用户名与密码不能为空"
	}

	QUERY := bson.M{"login": name, "ex.password": buildToken(passwd)}
	var u *Operator
	err := ds.se.DB(DB).C(C_OPERATOR).Find(QUERY).One(&u)
	if err != nil {
		return 400, "登录失败!"
	}

	err = ds.se.DB(DB).C(C_OPERATOR).UpdateId(u.Id, bson.M{"$set": bson.M{"lt": tick()}})
	chk(err)

	session.Set(ADMIN, u.Login)

	return web.Json(200, u)
}

func adminLogoutHandler(session sessions.Session) string {
	session.Delete(ADMIN)
	return ""
}

func adminMeHandler(web *Web, opSe *OpSession) (int, string) {
	return web.Json(200, opSe.me)
}

//type UpdatePasswdForm struct {
//	CurrentPasswd string `json:"currentPasswd" binding:"required"`
//	NewPasswd     string `json:"newPasswd" binding:"required"`
//}

func adminUpdatePasswdHandler(r *http.Request, opSe *OpSession, ds *Ds) (int, string) {

	currentPasswd := strings.TrimSpace(r.PostFormValue("currentPasswd"))
	newPasswd := strings.TrimSpace(r.PostFormValue("newPasswd"))

	if currentPasswd == "" || newPasswd == "" {
		return 400, "密码不能为空"
	}

	me := opSe.me
	if me.Ex.Password != buildToken(currentPasswd) {
		return 400, "当前密码错误!"
	}

	err := updateOperatorPassword(ds, me.Id, buildToken(newPasswd))
	if err != nil {
		log.Println("update op password err", me.Login, err)
		return 500, "更新出错!"
	}
	return 200, ""
}

func listOpHandler(r *http.Request, ds *Ds, web *Web) (int, string) {

	page, err := parseIntParam(r, "page", 1)
	if err != nil {
		return 400, err.Error()
	}
	size, err := parseIntParam(r, "size", 10)
	if err != nil {
		return 400, err.Error()
	}

	l, total, err := listOp(ds, nil, (page-1)*size, size)
	chk(err)

	j := J{
		"data":  l,
		"total": total,
		"page":  page,
		"size":  size,
	}
	return web.Json(200, j)
}

func createOpHandler(r *http.Request, ds *Ds, opse *OpSession) (int, string) {

	if opse.me.Login != "admin" {
		return 403, "没有权限"
	}

	login := strings.TrimSpace(r.PostFormValue("login"))
	passwd := strings.TrimSpace(r.PostFormValue("passwd"))

	op := &Operator{
		Id:    newId(),
		Login: login,

		Ex: &opEx{Password: buildToken(passwd)},
		Ct: tick(),
	}
	err := ds.se.DB(DB).C(C_OPERATOR).Insert(op)
	if dup(err) {
		return 400, "员工已存在!"
	}
	chk(err)
	return 200, ""
}

func showOpHandler(ds *Ds, web *Web, param martini.Params, op *Operator) (int, string) {
	return web.Json(200, op)
}

func delOpHandler(ds *Ds, web *Web, param martini.Params, op *Operator) (int, string) {
	if op.Login == "admin" {
		return 400, "admin不能删除"
	}

	err := delOperator(ds, op.Id)
	chk(err)
	return web.Json(200, "ok")
}

func checkOpLoginNameHandler(r *http.Request, ds *Ds, param martini.Params) (int, string) {

	name := strings.TrimSpace(r.PostFormValue("name"))

	if name == "" {
		return 400, "登录名不能为空!"
	}
	_, err := loadOperatorByLoginName(ds, name)

	if err != nil {
		if notFound(err) {
			return 200, ""
		} else {
			chk(err)
		}
	}

	return 200, "登录名已存在!"
}

func resetOpPasswordHandler(r *http.Request, ds *Ds) (int, string) {
	err := updateOperatorPassword(ds, bson.ObjectIdHex(strings.TrimSpace(r.PostFormValue("id"))), buildToken(strings.TrimSpace(r.PostFormValue("password"))))
	chk(err)
	return 200, ""
}
