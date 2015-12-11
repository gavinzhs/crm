package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func createOrgHandler(r *http.Request, web *Web, ds *Ds) (int, string) {
	parent := strings.TrimSpace(r.PostFormValue("parent"))
	name := strings.TrimSpace(r.PostFormValue("name"))
	tp := strings.TrimSpace(r.PostFormValue("tp"))
	addr := strings.TrimSpace(r.PostFormValue("addr"))
	owner := strings.TrimSpace(r.PostFormValue("owner"))
	mobile := strings.TrimSpace(r.PostFormValue("mobile"))
	buy := strings.TrimSpace(r.PostFormValue("buy"))
	memo := strings.TrimSpace(r.PostFormValue("memo"))

	rsp := &Rsp{
		Code: 0,
	}

	if name == "" || parent == "" || tp == "" {
		rsp.Data = "parameter required"
		return web.Json(200, rsp)
	}

	var parentId int
	var err error
	parentId, err = strconv.Atoi(parent)
	if err != nil {
		rsp.Data = fmt.Sprintf("bad parent: %s", parent)
		return web.Json(200, rsp)
	}

	var tpInt int
	tpInt, err = strconv.Atoi(tp)
	if err != nil {
		rsp.Data = fmt.Sprintf("bad tp: %s", tp)
		return web.Json(200, rsp)
	}

	if tpInt != ORG_TP_XIAOQU && tpInt != ORG_TP_LOUHAO && tpInt != ORG_TP_DANYUAN && tpInt != ORG_TP_MENPAIHAO {
		rsp.Data = fmt.Sprintf("bad tp: %d", tpInt)
		return web.Json(200, rsp)
	}

	var buyBool bool
	if buy != "" {
		buyBool, err = strconv.ParseBool(buy)
		if err != nil {
			rsp.Data = fmt.Sprintf("bad buy: %s", buy)
			return web.Json(200, rsp)
		}
	} else {
		buyBool = false
	}

	if parentId == 0 {
		root, err := loadOrgByQuery(ds.se, bson.M{"parent": parentId})
		if err != nil {
			rsp.Data = fmt.Sprintf("root org not found, error: %v", err)
			return web.Json(200, rsp)
		}
		parentId = root.Id
	}

	parentOrg, err := loadOrg(ds.se, parentId)
	if err != nil {
		if err == mgo.ErrNotFound {
			rsp.Data = "parent not found"
			return web.Json(200, rsp)
		}
		rsp.Data = fmt.Sprintf("get parent error: %v", err)
		return web.Json(200, rsp)
	}

	id, err := seq(ds.se, SEQ_ORG)
	if err != nil {
		rsp.Data = "generate id error"
		return web.Json(200, rsp)
	}
	org := &Org{
		Id:     id,
		Parent: parentId,
		Name:   name,
		Tp:     tpInt,
		Addr:   addr,
		Owner:  owner,
		Mobile: mobile,
		Buy:    buyBool,
		Memo:   memo,
		Ct:     tick(),
	}

	err = addOrg(ds.se, org)
	if err != nil {
		if dup(err) {
			rsp.Data = "duplicate err"
			return web.Json(200, rsp)
		}
		rsp.Data = fmt.Sprintf("create org error: %v", err)
		return web.Json(200, rsp)
	}

	// update parent info
	node := &Node{
		Id: org.Id,
		Tp: tpInt,
	}

	SPEC := bson.M{"$push": bson.M{"children": node}}
	err = ds.se.DB(DB).C(C_ORG).UpdateId(parentOrg.Id, SPEC)
	if err != nil {
		log.Printf("update parent error: %v", err)
		//delete
		err = delOrg(ds.se, org)
		chk(err)
		rsp.Data = fmt.Sprintf("update parent error: %v", err)
		return web.Json(200, rsp)
	}

	rsp.Code = 1
	rsp.Data = "create org successfully"
	return web.Json(200, rsp)
}

func portalShowOrgHandler(org *Org, web *Web, ds *Ds) (int, string) {
	prepareOrgForPortal(ds.se, org)

	return web.Json(200, org)
}

func updateOrgHandler(r *http.Request, org *Org, web *Web, ds *Ds) (int, string) {
	rsp := &Rsp{
		Code: 0,
	}

	spec := bson.M{}

	if s := strings.TrimSpace(r.PostFormValue("name")); s != "" {
		spec["name"] = s
	}

	if s := r.PostFormValue("addr"); s != "" {
		spec["addr"] = s
	}

	if s := r.PostFormValue("owner"); s != "" {
		spec["owner"] = s
	}

	if s := r.PostFormValue("mobile"); s != "" {
		spec["mobile"] = s
	}

	if s := r.PostFormValue("memo"); s != "" {
		spec["memo"] = s
	}

	if len(spec) == 0 {
		rsp.Data = fmt.Sprintf("parameter required")
		return web.Json(200, rsp)
	}

	err := ds.se.DB(DB).C(C_ORG).UpdateId(org.Id, bson.M{"$set": spec})
	if err != nil {
		rsp.Data = fmt.Sprintf("update org error: %v", err)
		return web.Json(200, rsp)
	}

	rsp.Code = 1
	rsp.Data = "update org successfully"
	return web.Json(200, rsp)
}

func delOrgHandler(org *Org, web *Web, ds *Ds) (int, string) {
	if org.Parent == 0 {
		return 403, "can't delete root org"
	}

	if err := delOrg(ds.se, org); err != nil {
		return 500, fmt.Sprintf("delete org error: %v", err)
	}

	return 200, ""
}
