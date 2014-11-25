package main

import (
  "html/template"
  "encoding/json"
  "net/http"
  "net"
  "time"
  "path"
  "math/rand"

  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
)

const HOMEPAGE = "http://hotrodup.com"

type App struct {
  IP  string
  Name string
  Runtime string
  Slug string
  Date    time.Time
}

func putApp(c appengine.Context, app *App) error {
  key := datastore.NewIncompleteKey(c, "App", nil)
  _, err := datastore.Put(c, key, app)
  return err
}

func getApp(c appengine.Context, slug string) (*App, error) {
  var apps []App
  q := datastore.NewQuery("App").Filter("Slug =", slug).Limit(1)
  if _, err := q.GetAll(c, &apps); err != nil {
    return &App{}, err
  }
  if len(apps) == 0 {
    return &App{}, nil
  }
  return &apps[0], nil
}

func init() {
  rand.Seed( time.Now().UTC().UnixNano())
  http.HandleFunc("/create", handleCreate)
  http.HandleFunc("/", handleBase)
}

var tpl = template.Must(template.ParseGlob("templates/*.html"))

func handleCreate(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
    return
  }

  ip := r.FormValue("ip")

  if net.ParseIP(ip) == nil {
    http.Error(w, "Not a valid IP address", http.StatusInternalServerError)
    return
  }

  c := appengine.NewContext(r)
  app := &App{
    IP: ip,
    Name: r.FormValue("name"),
    Runtime: r.FormValue("runtime"),
    Slug: randSeq(4),
    Date: time.Now(),
  }

  if err := putApp(c, app); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  js, err := json.Marshal(app)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.Write(js)
}

func handleBase(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
    return
  }
  if r.URL.Path == "/" {
    http.Redirect(w, r, HOMEPAGE, 301)
    return
  }

  slug := path.Base(rootDir(r.URL.Path))

  c := appengine.NewContext(r)

  app, err := getApp(c, slug)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if app == nil {
    http.NotFound(w, r)
    return
  }

  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  if err := tpl.ExecuteTemplate(w, "app.html", app); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}