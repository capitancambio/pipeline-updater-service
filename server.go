package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Release struct {
	Version    string
	Descriptor string
}

var URLDB = os.Getenv("URLDB")

func main() {

	sess, err := mgo.Dial(URLDB)
	if err != nil {
		panic(err)
	}
	col := sess.DB("pipeline-releases").C("releases")
	goji.Get("/:version", Get(col))
	goji.Post("/:version", Post(col))
	goji.Serve()

}

func Get(col *mgo.Collection) func(c web.C, w http.ResponseWriter, r *http.Request) {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		rel := Release{}
		version := c.URLParams["version"]

		err := col.Find(bson.M{"version": version}).One(&rel)
		if err != nil {
			log.Printf("%v", err)
			w.WriteHeader(404)
			return
		}
		log.Printf("Returning descriptor for %s", version)
		fmt.Fprintf(w, "%s", rel.Descriptor)

	}
}
func Post(col *mgo.Collection) func(c web.C, w http.ResponseWriter, r *http.Request) {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		//TODO:SOME SECURITY HERE WOULD BE NICE!
		version := c.URLParams["version"]
		buf := bytes.NewBufferString("")
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			log.Printf("%v", err)
			w.WriteHeader(500)
			w.Header().Set("X-Error", err.Error())
		}
		err = col.Insert(Release{Version: version, Descriptor: buf.String()})
		if err != nil {
			log.Printf("%v", err)
			w.WriteHeader(500)
			w.Header().Set("X-Error", err.Error())

		}
		log.Printf("Saved descriptor for %v", version)

	}
}
