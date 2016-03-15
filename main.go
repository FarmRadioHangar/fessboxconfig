package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/FarmRadioHangar/fessboxconfig/gsm"
	"github.com/gernest/hot"
	"github.com/gorilla/mux"
)

//Config holds configuration values for this application.
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
		tmp, err := ioutil.TempDir("", "fconf")
		if err != nil {
			log.Fatal(err)
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		go func(dir string) {
		END:
			for {
				select {
				case <-c:
					log.Println("removing ", dir)
					_ = os.RemoveAll(dir)
					break END
				}
			}
		}(tmp)
		cfg.AsteriskConfig = tmp
		err = copyFiles(tmp, "sample")
		if err != nil {
			log.Fatal(err)
		}
	}
	s := newServer(cfg)
	log.Printf(" starting server on  localhost:%d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), s))
}

//copyFiles copies files from src to dst, directories are ignored
func copyFiles(dst, src string) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fname := filepath.Join(src, file.Name())
		f, err := ioutil.ReadFile(fname)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dst, file.Name()), f, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

//newServer returns a http.Handler with all the routes for configuring supported
//devices registered.
func newServer(c *Config) http.Handler {
	s := mux.NewRouter()
	w := newWeb(c)
	s.HandleFunc("/device/dongle", w.Dongle).Methods("GET")
	s.HandleFunc("/device/dongle", w.UpdateDongle).Methods("POST")
	s.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(c.StaticDir))))
	s.HandleFunc("/", w.Home)
	return s
}

// web is the application struct, it defines all the handlers for this
// application as its methods.
//
// It is safe to use this in multiple goroutines
//
// This contains a a loaded hot template, so when run in development mode the
// templates will automatially be reloaded withouth the need to re run the
// application process. The auto reloading of templates is disabled in
// production.
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
	w.Header().Set("Content-Type", "application/json")

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

func (ww *web) UpdateDongle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ast := &gsm.Ast{}
	src := &bytes.Buffer{}
	enc := json.NewEncoder(w)
	_, err := io.Copy(src, r.Body)
	if err != nil {
		enc.Encode(&errMSG{Message: "trouble reading request body"})
		return
	}
	err = ast.LoadJSON(src.Bytes())
	if err != nil {
		enc.Encode(&errMSG{Message: "trouble loading request body"})
		return
	}
	fName := filepath.Join(ww.cfg.AsteriskConfig, "dongle.conf")
	info, err := os.Stat(fName)
	if err != nil {
		log.Println(err)
		enc.Encode(&errMSG{"trouble opening dongle configuration"})
		return
	}
	dst := &bytes.Buffer{}
	gsm.PrintAst(dst, ast)
	ioutil.WriteFile(fName, dst.Bytes(), info.Mode())
	io.Copy(w, src)
}
