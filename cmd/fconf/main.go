package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gernest/hot"
)

type Config struct {
	Port           int64  `json:"port"`
	Host           string `json:"host"`
	StaticDir      string `json:"static_dir"`
	TemplatesDir   string `json:"templates_dir"`
	AsteriskConfig string `json:"asterisk_config_dir"`
}

func defaultConfig() *Config {
	return &Config{
		Port:           8080,
		Host:           "",
		StaticDir:      "static",
		TemplatesDir:   "templates",
		AsteriskConfig: "/etc/asterisk",
	}
}

func main() {
	c := flag.String("c", "conf/config.json", "path to the configuration file")
	flag.Parse()
	b, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &Config{}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		log.Fatal(err)
	}
	s := newServer(cfg)
	log.Printf(" starting server on  localhost:%d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), s))
}

//newServer returns a http.Handler with all the routes for configuring supported
//devices registered.
func newServer(c *Config) http.Handler {
	s := http.NewServeMux()
	w := newWeb(c)
	s.HandleFunc("/", w.Home)
	s.HandleFunc("/device/dongle", w.Dongle)
	s.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(c.StaticDir))))
	return s
}

type web struct {
	cfg *Config
	tpl *hot.Template
}

func newWeb(cfg *Config) *web {
	w := &web{cfg: cfg}
	config := &hot.Config{
		Watch:          true,
		BaseName:       "fconf",
		Dir:            cfg.TemplatesDir,
		FilesExtension: []string{".tpl", ".html", ".tmpl"},
	}

	tpl, err := hot.New(config)
	if err != nil {
		panic(err)
	}
	w.tpl = tpl
	return w
}

func (ww *web) Home(w http.ResponseWriter, r *http.Request) {
	err := ww.tpl.Execute(w, "index.html", nil)
	if err != nil {
		log.Println(err)
	}
}

func (ww *web) Dongle(w http.ResponseWriter, r *http.Request) {
}
