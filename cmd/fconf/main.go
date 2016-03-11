package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FarmRadioHangar/fessboxconfig/gsm"
	"github.com/gernest/hot"
	"github.com/gorilla/mux"
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
	dev := flag.Bool("dev", false, "set true if running in dev mode")
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
	if *dev {
		cfg.AsteriskConfig = "sample"
	}
	s := newServer(cfg)
	log.Printf(" starting server on  localhost:%d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), s))
}

//newServer returns a http.Handler with all the routes for configuring supported
//devices registered.
func newServer(c *Config) http.Handler {
	s := mux.NewRouter()
	w := newWeb(c)
	s.HandleFunc("/device/dongle", w.Dongle).Methods("GET")
	s.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(c.StaticDir))))
	s.HandleFunc("/", w.Home)
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

type errMSG struct {
	Message string `json:"error"`
}

// Dongle implements http.HandleFunc for serving the dongle configuration values
// as a json object
func (ww *web) Dongle(w http.ResponseWriter, r *http.Request) {
	fName := filepath.Join(ww.cfg.AsteriskConfig, "dongle.conf")
	enc := json.NewEncoder(w)

	f, err := os.Open(fName)
	if err != nil {
		log.Println(err)
		enc.Encode(&errMSG{"trouble opening dongle configuration"})
		return
	}
	defer f.Close()
	p, err := gsm.NewParser(f)
	if err != nil {
		log.Println(err)
		enc.Encode(&errMSG{"trouble scanning dongle configuration"})
		return
	}
	ast, err := p.Parse()
	if err != nil {
		log.Println(err)
		enc.Encode(&errMSG{"trouble parsing dongle configuration"})
		return
	}
	err = ast.ToJSON(w)
	if err != nil {
		log.Println(err)
	}
}
