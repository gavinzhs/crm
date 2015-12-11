package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	DB = "crm"

	C_OPERATOR = "operator"
)

type Ds struct {
	se *mgo.Session
}

func (p *Ds) Copy() *Ds {
	return &Ds{se: p.se.Copy()}
}

func connect(db_con string) *mgo.Session {
	session, err := mgo.Dial(db_con)
	if err != nil {
		log.Fatal(err)
	}
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
