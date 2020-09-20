package main

import (
	"crypto/subtle"
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andreyvit/easyhttpserver"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Second), 5)

type User struct {
	AppTitle string `json:"app_title"`
	Password string `json:"password"`
}

type Config struct {
	Port     int             `json:"port"`
	Debug    bool            `json:"debug"`
	AppTitle string          `json:"app_title"`
	Users    map[string]User `json:"users"`
}

var config Config

var templ = template.Must(template.New("name").Parse(`
<!doctype html>
<html lang="ru" dir="ltr">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<meta http-equiv="Content-Language" content="ru" />
	<meta name="msapplication-TileColor" content="#2d89ef">
	<meta name="theme-color" content="#4188c9">
	<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/>
	<meta name="apple-mobile-web-app-capable" content="yes">
	<meta name="mobile-web-app-capable" content="yes">
	<meta name="HandheldFriendly" content="True">
	<meta name="MobileOptimized" content="320">
	<!-- <link rel="icon" href="./favicon.ico" type="image/x-icon"/> -->
	<!-- <link rel="shortcut icon" type="image/x-icon" href="/favicon.ico" /> -->
	<title>{{.AppTitle}}</title>
</head>
<title>{{.AppTitle}}</title>
<style>
	html {
		font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
		font-size: 150%;
	}
	body {
		margin: 0 auto;
		max-width: 50em;
		line-height: 1.5;
		padding: 4em 1em;
	}
</style>
<body>
	{{range .Messages}}
	<p>[{{.TimeStr}}] <b>{{.Sender}}:</b> {{.Text}}</p>
	{{end}}
</body>
</html>
`))

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime)

	var configFile string
	flag.StringVar(&configFile, "f", "", "configuration file path (.json)")
	flag.Parse()

	if configFile == "" {
		log.Fatalf("** Config file not specified. Pass -f path/to/config.json")
	}
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("** Cannot read config file %v: %v", configFile, err)
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatalf("** Cannot parse config file %v: %v", configFile, err)
	}
	if len(config.Users) == 0 {
		log.Fatalf("** No users configured in config file")
	}

	err = LoadMessages(func(msg Message) {
		if config.Debug && isCode(msg.Text) {
			log.Printf("%v\n", msg)
		}
	})
	if err != nil {
		log.Fatalf("** Cannot load messages: %v", err)
	}

	sopt := easyhttpserver.Options{
		DefaultDevPort:          7000,
		GracefulShutdownTimeout: 2 * time.Second,
		Port:                    config.Port,
	}

	srv, err := easyhttpserver.Start(http.HandlerFunc(HandleRequest), sopt)
	if err != nil {
		log.Fatalf("** %v", err)
	}
	easyhttpserver.InterceptShutdownSignals(srv.Shutdown)

	log.Printf("SMS Code Remote Access server running at %s", strings.Join(srv.Endpoints(), ", "))

	err = srv.Wait()
	if err != nil {
		log.Fatalf("** %v", err)
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// rate limit
	now := time.Now()
	rsrv := limiter.ReserveN(now, 1)
	delay := rsrv.DelayFrom(now)
	if delay > 3*time.Second {
		rsrv.CancelAt(now)
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	time.Sleep(delay)

	// auth
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Login Required\", charset=\"utf-8\"")
	user, ok := authenticate(r.BasicAuth())
	if !ok {
		http.Error(w, "Login Required", http.StatusUnauthorized)
		return
	}

	showAll := false
	if r.FormValue("all") == "1" {
		showAll = true
	}

	var msgs []Message
	err := LoadMessages(func(msg Message) {
		if showAll || (len(msgs) < 10 && isCode(msg.Text)) {
			msgs = append(msgs, msg)
		}
	})
	if err != nil {
		log.Printf("ERROR: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appTitle := config.AppTitle
	if user.AppTitle != "" {
		appTitle = user.AppTitle
	}

	w.Header().Set("Content-Type", "text/html, charset=\"utf-8\"")
	err = templ.Execute(w, struct {
		Messages []Message
		AppTitle string
	}{
		Messages: msgs,
		AppTitle: appTitle,
	})
	if err != nil {
		log.Printf("ERROR: rendering template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func authenticate(u, pw string, ok bool) (User, bool) {
	if !ok {
		return User{}, false
	}
	if user, ok := config.Users[u]; ok && user.Password != "" {
		if 1 == subtle.ConstantTimeCompare([]byte(pw), []byte(user.Password)) {
			return user, true
		}
	}
	return User{}, false
}
