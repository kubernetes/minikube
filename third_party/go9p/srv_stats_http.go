//go:build httpstats

package go9p

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

var mux sync.RWMutex
var stat map[string]http.Handler
var httponce sync.Once

func register(s string, h http.Handler) {
	mux.Lock()
	if stat == nil {
		stat = make(map[string]http.Handler)
	}

	if h == nil {
		delete(stat, s)
	} else {
		stat[s] = h
	}
	mux.Unlock()
}
func (srv *Srv) statsRegister() {
	httponce.Do(func() {
		http.HandleFunc("/go9p/", StatsHandler)
		go http.ListenAndServe(":6060", nil)
	})

	register("/go9p/srv/"+srv.Id, srv)
}

func (srv *Srv) statsUnregister() {
	register("/go9p/srv/"+srv.Id, nil)
}

func (srv *Srv) ServeHTTP(c http.ResponseWriter, r *http.Request) {
	io.WriteString(c, fmt.Sprintf("<html><body><h1>Server %s</h1>", srv.Id))
	defer io.WriteString(c, "</body></html>")

	// connections
	io.WriteString(c, "<h2>Connections</h2><p>")
	srv.Lock()
	defer srv.Unlock()
	if len(srv.conns) == 0 {
		io.WriteString(c, "none")
		return
	}

	for _, conn := range srv.conns {
		io.WriteString(c, fmt.Sprintf("<a href='/go9p/srv/%s/conn/%s'>%s</a><br>", srv.Id, conn.Id, conn.Id))
	}
}

func (conn *Conn) statsRegister() {
	register("/go9p/srv/"+conn.Srv.Id+"/conn/"+conn.Id, conn)
}

func (conn *Conn) statsUnregister() {
	register("/go9p/srv/"+conn.Srv.Id+"/conn/"+conn.Id, nil)
}

func (conn *Conn) ServeHTTP(c http.ResponseWriter, r *http.Request) {
	io.WriteString(c, fmt.Sprintf("<html><body><h1>Connection %s/%s</h1>", conn.Srv.Id, conn.Id))
	defer io.WriteString(c, "</body></html>")

	// statistics
	conn.Lock()
	io.WriteString(c, fmt.Sprintf("<p>Number of processed requests: %d", conn.nreqs))
	io.WriteString(c, fmt.Sprintf("<br>Sent %v bytes", conn.rsz))
	io.WriteString(c, fmt.Sprintf("<br>Received %v bytes", conn.tsz))
	io.WriteString(c, fmt.Sprintf("<br>Pending requests: %d max %d", conn.npend, conn.maxpend))
	io.WriteString(c, fmt.Sprintf("<br>Number of reads: %d", conn.nreads))
	io.WriteString(c, fmt.Sprintf("<br>Number of writes: %d", conn.nwrites))
	conn.Unlock()

	// fcalls
	if conn.Debuglevel&DbgLogFcalls != 0 {
		fs := conn.Srv.Log.Filter(conn, DbgLogFcalls)
		io.WriteString(c, fmt.Sprintf("<h2>Last %d 9P messages</h2>", len(fs)))
		for i, l := range fs {
			fc := l.Data.(*Fcall)
			if fc.Type == 0 {
				continue
			}

			lbl := ""
			if fc.Type%2 == 0 {
				// try to find the response for the T message
				for j := i + 1; j < len(fs); j++ {
					rc := fs[j].Data.(*Fcall)
					if rc.Tag == fc.Tag {
						lbl = fmt.Sprintf("<a href='#fc%d'>&#10164;</a>", j)
						break
					}
				}
			} else {
				// try to find the request for the R message
				for j := i - 1; j >= 0; j-- {
					tc := fs[j].Data.(*Fcall)
					if tc.Tag == fc.Tag {
						lbl = fmt.Sprintf("<a href='#fc%d'>&#10166;</a>", j)
						break
					}
				}
			}

			io.WriteString(c, fmt.Sprintf("<br id='fc%d'>%d: %s%s", i, i, fc, lbl))
		}
	}
}

func StatsHandler(c http.ResponseWriter, r *http.Request) {
	mux.RLock()
	if v, ok := stat[r.URL.Path]; ok {
		v.ServeHTTP(c, r)
	} else if r.URL.Path == "/go9p/" {
		io.WriteString(c, fmt.Sprintf("<html><body><br><h1>On offer: </h1><br>"))
		for v := range stat {
			io.WriteString(c, fmt.Sprintf("<a href='%s'>%s</a><br>", v, v))
		}
		io.WriteString(c, "</body></html>")
	}
	mux.RUnlock()
}
