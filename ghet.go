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

func checkError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	var uri = flag.String("u", "", "uri")
	var output = flag.String("o", "", "output")
	var setup = flag.Bool("setup", false, "setup ghet")
	var list = flag.Bool("list", false, "list items")
	var update = flag.Bool("update", false, "update an item")
	flag.Parse()

	confDir := path.Join(os.Getenv("HOME"), ".ghet")
	confPath := path.Join(confDir, "ghet.conf")

	// Setup
	if *setup {
		if _, err := os.Stat(confDir); os.IsNotExist(err) {
			err = os.MkdirAll(confDir, os.ModeDir|0775)
			checkError(err)
		}
		c := config.NewDefault()
		var gat string
		log.Printf("github access token (https://github.com/settings/tokens/new): ")
		_, err := fmt.Scanln(&gat)
		checkError(err)
		c.AddSection("github")
		c.AddOption("github", "access_token", gat)
		c.WriteFile(confPath, 0644, "ghet")
		return
	}

	// Load config
	conf, err := config.ReadDefault(confPath)
	checkError(err)

	// Load db
	dbPath := path.Join(confDir, "db")
	db, err := leveldb.OpenFile(dbPath, nil)
	defer db.Close()

	// List items
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

	// Get url
	var u *url.URL
	if *update {
		if *output == "" {
			log.Fatalf("specify file to update with -o")
		}
		surl, err := db.Get([]byte(*output), nil)
		checkError(err)
		u, err = url.ParseRequestURI(string(surl))
	} else {
		if *uri == "" {
			log.Fatalf("no uri specified")
			return
		}
		u, err = url.ParseRequestURI(*uri)
	}
	checkError(err)

	// Download/update
	var (
		name    string
		content []byte
	)
	if u.Host != "github.com" {
		log.Fatalf("not a github uri")
	}
	name, content = fetch(conf, u)
	if *output != "" {
		name = *output
	}
	absPath, err := filepath.Abs(name)
	checkError(err)
	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0777)
	defer f.Close()
	_, err = f.Write(content)
	checkError(err)
	err = db.Put([]byte(absPath), []byte(*uri), nil)
	checkError(err)
	log.Printf("%s -> %s", u.String(), absPath)
}

type GithubRepoContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func fetch(conf *config.Config, u *url.URL) (string, []byte) {
	tok, err := conf.String("github", "access_token")
	checkError(err)
	ght := &oauth.Transport{Token: &oauth.Token{AccessToken: tok}}
	ghc := github.NewClient(ght.Client())
	us := strings.Split(u.Path, "/")
	or := strings.Trim(strings.Join(us[:3], "/"), "/")
	p := strings.Join(us[5:], "/")
	ru := strings.Join([]string{"repos", or, "contents", p}, "/")
	req, err := ghc.NewRequest("GET", ru, nil)
	checkError(err)
	rc := GithubRepoContent{}
	_, err = ghc.Do(req, &rc)
	checkError(err)
	dec, err := base64.StdEncoding.DecodeString(rc.Content)
	checkError(err)
	return rc.Name, dec
}
