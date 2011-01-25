package main

import (
	"flag"
	"github.com/nf/stat"
	"http"
	"json"
	"os"
	"rpc"
	"sync"
	"time"
)

var (
	listenAddr = flag.String("http", ":8090", "HTTP listen port")
	maxLen     = flag.Int("max", 60, "max points to retain")
)

type Server map[string][][2]int64

func (s Server) Update(args *stat.Point, r *struct{}) os.Error {
	lock.Lock()
	t := (time.Nanoseconds() - start) / 100e6
	k := args.Process + " " + args.Series
	s[k] = append(s[k], [2]int64{t, args.Value})
	if len(s[k]) > *maxLen {
		s[k] = s[k][len(s[k])-*maxLen:]
	}
	lock.Unlock()
	return nil
}

var (
	server = make(Server)
	start  = time.Nanoseconds()
	lock   sync.Mutex
)

func main() {
	flag.Parse()
	rpc.Register(server)
	rpc.HandleHTTP()
	http.HandleFunc("/", Static)
	http.HandleFunc("/get", Get)
	http.ListenAndServe(*listenAddr, nil)
}

func Get(w http.ResponseWriter, r *http.Request) {
	w.SetHeader("Content-Type", "application/json")
	lock.Lock()
	e := json.NewEncoder(w)
	e.Encode(server)
	lock.Unlock()
}

func Static(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	if filename == "" {
		filename = "index.html"
	} else if filename[:6] != "flotr/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/"+filename)
}
