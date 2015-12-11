package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	ORG_TP_XIAOQU = iota
	ORG_TP_LOUHAO
	ORG_TP_DANYUAN
	ORG_TP_MENPAIHAO
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

type Org struct {
	Id       int     `bson:"_id" json:"id"`
	Parent   int     `bson:"parent" json:"parent"`
	Name     string  `bson:"name" json:"name"`
	Children []*Node `bson:"children" json:"children"`
	Ct       int64   `bson:"ct" json:"ct"`
	Tp       int     `bson:"tp" json:"tp"`
	Addr     string  `bson:"addr,omitempty" json:"addr"`
	Owner    string  `bson:"owner,omitempty" json:"owner"`
	Mobile   string  `bson:"mobile,omitempty" json:"mobile"`
	Buy      bool    `bson:"buy" json:"buy"`
	Memo     string  `bson:"memo,omitempty" json:"memo"`
}

/**
 * 用来返回组织结构的children，可以是用户，也可以是组织结构
 */
type Node struct {
	Id     int    `json:"id"`
	Name   string `bson:"-" json:"name"`
	Tp     int    `bson:"-" json:"tp"`
	IsLeaf bool   `bson:"-" json:"isLeaf"` //如果是叶子结点，则表示没有children
	Addr   string `bson:"-" json:"addr"`
	Owner  string `bson:"-" json:"owner"`
	Mobile string `bson:"-" json:"mobile"`
	Buy    bool   `bson:"-" json:"buy"`
	Memo   string `bson:"-" json:"memo"`
}

type Seq struct {
	Id   bson.ObjectId `bson:"_id"`
	Name string
	N    int
}

func ensureIndex(session *mgo.Session) {
	idx := mgo.Index{Key: []string{"login"}, Unique: true}
	err := session.DB(DB).C(C_OPERATOR).EnsureIndex(idx)
	chk(err)

    idx = mgo.Index{Key: []string{"name", "parent"}, Unique: true}
    err = session.DB(DB).C(C_ORG).EnsureIndex(idx)
    chk(err)

	//add root worldOrg if not exist
	_, err = loadOrgByQuery(session, bson.M{"parent": 0})
	if notFound(err) {
		n, err := seq(session, SEQ_ORG)
		if err != nil {
			log.Fatalf("seq err : %v", err)
		}
		org := &Org{
			Id:   n,
			Name: "安哥地盘",
			Tp:   -1,
			Ct:   tick(),
		}
		err = addOrg(session, org)
		chk(err)
	} else {
		chk(err)
	}
}
