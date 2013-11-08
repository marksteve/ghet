package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/robfig/config"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
  "text/tabwriter"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	var confDir string
	if _, err := os.Stat("dwn.conf"); err == nil {
		confDir = "./"
	} else {
		confDir = path.Join(os.Getenv("HOME"), ".dwn")
	}
	conf, err := config.ReadDefault(path.Join(confDir, "dwn.conf"))
	if err != nil {
		log.Fatalf("%s", err)
	}
	dbPath := path.Join(confDir, "db")
	db, err := leveldb.OpenFile(dbPath, nil)
	defer db.Close()
  var uri = flag.String("u", "", "uri to dwn")
  var output = flag.String("o", "", "dwn to output")
	var list = flag.Bool("list", false, "list dwn'd items")
	var update = flag.Bool("update", false, "update dwn'd item")
	flag.Parse()
  if *list {
    w := &tabwriter.Writer{}
    w.Init(os.Stdout, 0, 8, 1, ' ', 0)
    fmt.Fprintln(w, "Path\tURI")
    iter := db.NewIterator(nil)
    for iter.Next() {
      key := iter.Key()
      value := iter.Value()
      fmt.Fprintf(w, "%s\t%s\n", key, value)
    }
    w.Flush()
    return
  }
  var u *url.URL
  if *update {
    if *output == "" {
      log.Fatalf("specify file to update with -o")
    }
    surl, err := db.Get([]byte(*output), nil)
    if err != nil {
      log.Fatalf("%s", err)
    }
    u, err = url.ParseRequestURI(string(surl))
  } else {
    if *uri == "" {
      log.Fatalf("no uri specified")
      return
    }
    u, err = url.ParseRequestURI(*uri)
    if err != nil {
      log.Fatalf("%s", err)
    }
  }
	var (
		name    string
		content []byte
	)
	switch u.Host {
	case "github.com":
		name, content = Github(conf, u)
	}
	if *output != "" {
		name = *output
	}
	absPath, err := filepath.Abs(name)
	if err != nil {
		log.Fatalf("%s", err)
	}
	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0777)
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = db.Put([]byte(absPath), []byte(*uri), nil)
	if err != nil {
		log.Fatal("%s", err)
	}
  log.Printf("%s -> %s", u.String(), absPath)
}

type GithubRepoContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func Github(conf *config.Config, u *url.URL) (string, []byte) {
	tok, err := conf.String("github", "access_token")
	if err != nil {
		log.Fatalf("%s", err)
	}
	ght := &oauth.Transport{Token: &oauth.Token{AccessToken: tok}}
	ghc := github.NewClient(ght.Client())
	us := strings.Split(u.Path, "/")
	or := strings.Trim(strings.Join(us[:3], "/"), "/")
	p := strings.Join(us[5:], "/")
	ru := strings.Join([]string{"repos", or, "contents", p}, "/")
	req, err := ghc.NewRequest("GET", ru, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}
	rc := GithubRepoContent{}
	_, err = ghc.Do(req, &rc)
	if err != nil {
		log.Fatalf("%s", err)
	}
	dec, err := base64.StdEncoding.DecodeString(rc.Content)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return rc.Name, dec
}
