package main

import (
	"github.com/go-martini/martini"
	"github.com/quexer/sessions"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strconv"
)

type AdminLoginForm struct {
	Name   string `json:"name" binding:"required"`
	Passwd string `json:"passwd" binding:"required"`
}

func adminLoginHandler(form AdminLoginForm, web *Web, ds *Ds, session sessions.Session) (int, string) {
	QUERY := bson.M{"login": form.Name, "ex.password": buildToken(form.Passwd)}
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

type UpdatePasswdForm struct {
	CurrentPasswd string `json:"currentPasswd" binding:"required"`
	NewPasswd     string `json:"newPasswd" binding:"required"`
}

func adminUpdatePasswdHandler(form UpdatePasswdForm, opSe *OpSession, ds *Ds) (int, string) {
	me := opSe.me
	if me.Ex.Password != buildToken(form.CurrentPasswd) {
		return 400, "当前密码错误!"
	}

	err := updateOperatorPassword(ds, me.Id, buildToken(form.NewPasswd))
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

type OpForm struct {
	Login  string `json:"login" binding:"required"`
	Passwd string `json:"passwd" binding:"required"`
}

func createOpHandler(form OpForm, ds *Ds) (int, string) {
	op := &Operator{
		Id:    newId(),
		Login: form.Login,

		Ex: &opEx{Password: buildToken(form.Passwd)},
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

type NameForm struct {
	Name string `binding:"required"`
}

func checkOpLoginNameHandler(form NameForm, ds *Ds, param martini.Params) (int, string) {

	if form.Name == "" {
		return 400, "登录名不能为空!"
	}
	_, err := loadOperatorByLoginName(ds, form.Name)

	if err != nil {
		if notFound(err) {
			return 200, ""
		} else {
			chk(err)
		}
	}

	return 200, "登录名已存在!"
}

type UpdateVoForm struct {
	Id  bson.ObjectId `json:"id" binding:"required"`
	Key string        `json:"key" binding:"requried"`
	Val string        `json:"val"`
}

func updateOpHandler(form UpdateVoForm, ds *Ds, op *Operator) (int, string) {
	switch form.Key {
	case "name", "email", "mobile":
		SPEC := bson.M{"mt": tick()}
		SPEC[form.Key] = form.Val
		err := ds.se.DB(DB).C(C_OPERATOR).UpdateId(op.Id, bson.M{"$set": SPEC})
		chk(err)
	case "system":
		sys, err := strconv.Atoi(form.Val)
		chk(err)
		SPEC := bson.M{"mt": tick()}
		SPEC[form.Key] = sys
		err = ds.se.DB(DB).C(C_OPERATOR).UpdateId(op.Id, bson.M{"$set": SPEC})
		chk(err)
	default:
		return 400, "[warning] unkown edit key: " + form.Key
	}
	return 200, ""
}

type ResetOpPasswordForm struct {
	Id       bson.ObjectId `json:"id" binding:"required"`
	Password string        `json:"password" binding:"required"`
}

func resetOpPasswordHandler(form ResetOpPasswordForm, ds *Ds) (int, string) {
	err := updateOperatorPassword(ds, form.Id, buildToken(form.Password))
	chk(err)
	return 200, ""
}
