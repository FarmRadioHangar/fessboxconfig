package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type DeviceConfig interface {
	Name() string
	LoadJSON(io.Reader) error
	Save() error
	ToJSON(io.Writer) error
}

type Config struct {
	Port           int64  `json:"port"`
	Host           string `json:"host"`
	StaticDir      string `json:"static_dir"`
	IndexFile      string `json:"index_file"`
	AsteriskConfig string `json:"asterisk_config_dir"`
}

func defaultConfig() *Config {
	return &Config{
		Port:           8080,
		Host:           "",
		StaticDir:      "static",
		IndexFile:      "index.html",
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
	s := http.NewServeMux()
	w := newWeb(cfg)
	s.HandleFunc("/", w.Home)
	s.HandleFunc("/device/dongle", w.Dongle)
	s.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	log.Printf(" starting server on  localhost:%d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), s))
}

type web struct {
	cfg     *Config
	homeTpl *template.Template
}

func newWeb(cfg *Config) *web {
	w := &web{cfg: cfg}
	b, err := ioutil.ReadFile(cfg.IndexFile)
	if err != nil {
		log.Fatal(err)
	}
	tpl, err := template.New("index").Parse(string(b))
	if err != nil {
		log.Fatal(err)
	}
	w.homeTpl = tpl
	return w
}

func (ww *web) Home(w http.ResponseWriter, r *http.Request) {
	err := ww.homeTpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		log.Println(err)
	}
}

func (ww *web) Dongle(w http.ResponseWriter, r *http.Request) {
}
