package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type J map[string]interface{}

type opEx struct {
	Password string
}

//管理员
type Operator struct {
	Id    bson.ObjectId `bson:"_id" json:"id"`
	Login string        `json:"login"` //登录名
	Ct    int64         `json:"ct"`    //创建时间
	Mt    int64         `json:"mt"`    //更后一次更新时间
	Lt    int64         `json:"lt"`    //最后一次登录时间
	Ex    *opEx         `json:"-"`
}

func ensureIndex(db *mgo.Session) {
	idx := mgo.Index{Key: []string{"login"}, Unique: true}
	err := db.DB(DB).C(C_OPERATOR).EnsureIndex(idx)
	chk(err)
}
