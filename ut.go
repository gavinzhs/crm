package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"net/http"
	"strconv"
	"time"
)

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func dup(err error) bool {
	return mgo.IsDup(err)
}

func notFound(err error) bool {
	return err == mgo.ErrNotFound
}

func nano() int64 {
	return time.Now().UnixNano()
}

func tick() int64 {
	return nano() / 1e6
}

func newId() bson.ObjectId {
	return bson.NewObjectId()
}

func buildToken(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	io.WriteString(h, SALT)
	return fmt.Sprintf("%x", h.Sum(nil))
}

/*
 web utilities
*/
type Web struct {
	w http.ResponseWriter
}

func (p *Web) Json(code int, data interface{}) (int, string) {
	b, err := json.Marshal(data)
	chk(err)
	p.w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	return code, string(b)
}

func (p *Web) Code(code int) (int, string) {
	return code, http.StatusText(code)
}

type Rsp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func parseIntParam(r *http.Request, name string, defaultVal int) (int, error) {
	val := defaultVal
	if s := r.FormValue(name); s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("bad %s: %s", name, s)
		}
		val = i
	}
	return val, nil
}
