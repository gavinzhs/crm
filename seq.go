/**
 *
 */
package main

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	SEQ_ORG = "org"
)

func seq(session *mgo.Session, name string) (int, error) {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"n": 1}},
		ReturnNew: true,
	}
	n := new(Seq)
	_, err := session.DB(DB).C(C_SEQ).Find(bson.M{"name": name}).Apply(change, n)
	//	log.Println(err, n)
	if err != nil {
		return 0, err
	}
	return n.N, nil
}

func initSeq(session *mgo.Session) error {
	log.Println("初始化计数器")
	var n Seq
	c := session.DB(DB).C(C_SEQ)
	INIT_NAMES := []string{SEQ_ORG}
	for _, name := range INIT_NAMES {
		spec := bson.M{"name": name}

		err := c.Find(spec).One(&n)
		if err == mgo.ErrNotFound {
			_, err = c.Upsert(spec, bson.M{"$inc": bson.M{"n": 100000}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
