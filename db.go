package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	DB = "crm"

	C_OPERATOR = "operator"
	C_ORG      = "org"
	C_SEQ      = "seq"
)

type Ds struct {
	se *mgo.Session
}

func (p *Ds) Copy() *Ds {
	return &Ds{se: p.se.Copy()}
}

func connect(db_con string) *mgo.Session {
	log.Printf("开始连接:%s", db_con)
	session, err := mgo.Dial(db_con)
	if err != nil {
		log.Println("连接失败")
		log.Fatal(err)
	}
	log.Println("连接成功")
	return session
}

func initData(se *mgo.Session) {
	admin := &Operator{Id: newId(), Login: "admin",
		Ex: &opEx{Password: buildToken("123456")},
		Ct: tick(),
	}
	if err := se.DB(DB).C(C_OPERATOR).Insert(admin); err != nil {
		if !dup(err) {
			log.Fatal(err)
		}
	}
}

func loadOperatorByLoginName(ds *Ds, name string) (*Operator, error) {
	var op *Operator
	err := ds.se.DB(DB).C(C_OPERATOR).Find(bson.M{"login": name}).One(&op)
	return op, err
}

func loadOperator(ds *Ds, id bson.ObjectId) (*Operator, error) {
	var op *Operator
	err := ds.se.DB(DB).C(C_OPERATOR).FindId(id).One(&op)
	return op, err
}

func listOp(ds *Ds, query bson.M, skip int, limit int) ([]*Operator, int, error) {
	l := []*Operator{}
	Q := ds.se.DB(DB).C(C_OPERATOR).Find(query).Sort("-ct")
	total, err := Q.Count()
	if err != nil {
		return nil, 0, err
	}

	if err := Q.Skip(skip).Limit(limit).All(&l); err != nil {
		return nil, 0, err
	}

	return l, total, nil

}

func updateOperatorPassword(ds *Ds, id bson.ObjectId, password string) error {
	SPEC := bson.M{"$set": bson.M{"ex.password": password}}
	err := ds.se.DB(DB).C(C_OPERATOR).UpdateId(id, SPEC)
	return err
}

func delOperator(ds *Ds, id bson.ObjectId) error {
	return ds.se.DB(DB).C(C_OPERATOR).RemoveId(id)
}

//Org
func loadOrg(session *mgo.Session, id int) (*Org, error) {
	var org *Org
	err := session.DB(DB).C(C_ORG).FindId(id).One(&org)
	return org, err
}

func loadOrgByQuery(session *mgo.Session, query bson.M) (*Org, error) {
	var org *Org
	if err := session.DB(DB).C(C_ORG).Find(query).One(&org); err != nil {
		return nil, err
	}
	return org, nil
}

func loadOrgs(session *mgo.Session, ids []int) ([]*Org, error) {
	l := []*Org{}
	err := session.DB(DB).C(C_ORG).Find(bson.M{"_id": bson.M{"$in": ids}}).All(&l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func listOrgByQuery(session *mgo.Session, query bson.M) ([]*Org, error) {
	l := []*Org{}
	err := session.DB(DB).C(C_ORG).Find(query).All(&l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func listOrg(ds *Ds, query bson.M, skip int, limit int) ([]*Org, int, error) {
	l := []*Org{}
	Q := ds.se.DB(DB).C(C_ORG).Find(query).Sort("name")
	total, err := Q.Count()
	if err != nil {
		return nil, 0, err
	}

	if err := Q.Skip(skip).Limit(limit).All(&l); err != nil {
		return nil, 0, err
	}

	return l, total, nil

}

func addOrg(session *mgo.Session, org *Org) error {
	return session.DB(DB).C(C_ORG).Insert(org)
}

func updateOrg(session *mgo.Session, org *Org) error {
	return session.DB(DB).C(C_ORG).UpdateId(org.Id, org)
}

func delOrg(session *mgo.Session, org *Org) error {
	node := &Node{
		Id: org.Id,
		Tp: org.Tp,
	}

	//将该节点从父节点中删除
	SPEC := bson.M{"$pull": bson.M{"children": node}}
	if err := session.DB(DB).C(C_ORG).UpdateId(org.Parent, SPEC); err != nil {
		return err
	}

	//todo 没有处理下级成员  按说应该是删除  不应该直接升级为上一级菜单子类
	return session.DB(DB).C(C_ORG).RemoveId(org.Id)
}

func prepareOrgForPortal(session *mgo.Session, org *Org) error {
	for _, child := range org.Children {
		log.Println("child id:", child.Id)
		o, err := loadOrg(session, child.Id)
		if err != nil {
			return err
		}
		child.Name = o.Name
		child.Addr = o.Addr
		child.Owner = o.Owner
		child.Mobile = o.Mobile
		child.Buy = o.Buy
		child.Memo = o.Memo

		child.IsLeaf = (len(o.Children) == 0)
	}

	return nil
}
