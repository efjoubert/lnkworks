package lnksworks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Route struct {
	rootpath string

	hndlefunc func(svr *Server, rt *Route, root string, path string, mimetype string, w http.ResponseWriter, r *http.Request, active *ActiveProcessor)
}

func (rt *Route) cleanupRoute() {
	if rt.hndlefunc != nil {
		rt.hndlefunc = nil
	}

}

func retrieveRs(rt *Route, root string, path string, retrievedRs map[string]io.ReadSeeker) (io.ReadSeeker, error) {
	if root == "" {
		root = rt.rootpath
	}
	if rsfound, rsfoundok := retrievedRs[path]; rsfoundok {
		if rsf, rsfok := rsfound.(*os.File); rsfok {
			_, rsfseekerr := rsf.Seek(0, 0)
			return rsf, rsfseekerr
		} else if rsemcur, rsemcurok := rsfound.(*ReadWriteCursor); rsemcurok {
			_, rscurseekerr := rsemcur.Seek(0, 0)
			return rsemcur, rscurseekerr
		}
	} else {
		if embres := EmbededResourceByPath(path); embres != nil {
			embrescur := embres.ReadWriteCursor(true)
			_, embrescurseekerr := embrescur.Seek(0, 0)
			retrievedRs[path] = embrescur
			return embrescur, embrescurseekerr
		} else if f, ferr := os.Open(root + path); ferr == nil {
			_, fseekerr := f.Seek(0, 0)
			retrievedRs[path] = f
			return f, fseekerr
		} else if ferr != nil {
			return nil, ferr
		}
	}
	return nil, nil
}

func (rt *Route) ServeContent(path string, w http.ResponseWriter, r *http.Request, active *ActiveProcessor, retrievedRs map[string]io.ReadSeeker) {
	rs, rserr := retrieveRs(rt, "", path, retrievedRs)
	if rs != nil && rserr == nil {
		if active == nil {
			http.ServeContent(w, r, path, time.Now(), rs)
		} else {
			active.Process(rs, rt.rootpath, path, func(root string, path string) (rsfound io.ReadSeeker, rsfounderr error) {
				rsfound, rsfounderr = retrieveRs(rt, root, path, retrievedRs)
				return rsfound, rsfounderr
			})
		}
	} else if rserr != nil {
		w.Write([]byte(rserr.Error()))
	}
	if retrievedRs != nil {
		if len(retrievedRs) > 0 {
			for rspath, rs := range retrievedRs {
				if rsf, rsfok := rs.(*os.File); rsfok {
					rsf.Close()
					rsf = nil
				} else if rscur, rscurok := rs.(*ReadWriteCursor); rscurok {
					rscur.Close()
					rscur = nil
				}
				delete(retrievedRs, rspath)
			}
		}
		retrievedRs = nil
	}
}

type Router struct {
	mappedRoutes map[string]*Route
	rlock        *sync.RWMutex
}

func RegisterRoute(path string, rootpath string, hndlefunc ...func(svr *Server, rt *Route, root string, path string, mimetype string, w http.ResponseWriter, r *http.Request, active *ActiveProcessor)) {
	if !strings.HasSuffix(rootpath, "/") {
		if rootpath == "" {
			rootpath = "./"
		} else {
			rootpath = rootpath + "/"
		}
	}
	routes.rlock.RLock()
	defer routes.rlock.RUnlock()
	if _, okpath := routes.mappedRoutes[path]; !okpath {
		if len(hndlefunc) == 1 {
			routes.mappedRoutes[path] = &Route{rootpath: rootpath, hndlefunc: hndlefunc[0]}
		} else if len(hndlefunc) == 0 {
			routes.mappedRoutes[path] = &Route{rootpath: rootpath, hndlefunc: func(svr *Server, rt *Route, root string, path string, mimetype string, w http.ResponseWriter, r *http.Request, active *ActiveProcessor) {
				var retrievedRs map[string]io.ReadSeeker = make(map[string]io.ReadSeeker)
				rt.ServeContent(path, w, r, active, retrievedRs)
				if retrievedRs != nil {
					if len(retrievedRs) > 0 {
						for rspath, rs := range retrievedRs {
							if rsf, rsfok := rs.(*os.File); rsfok {
								rsf.Close()
								rsf = nil
							} else if rscur, rscurok := rs.(*ReadWriteCursor); rscurok {
								rscur.Close()
								rscur = nil
							}
							delete(retrievedRs, rspath)
						}
					}
					retrievedRs = nil
				}
			}}
		}
	}
}

func loadParametersFromHttpRequest(params *Parameters, r *http.Request) {
	if err := r.ParseMultipartForm(0); err == nil {
		if r.MultipartForm != nil {
			for pname, pvalue := range r.MultipartForm.Value {
				params.SetParameter(pname, false, pvalue...)
			}
			for pname, pfile := range r.MultipartForm.File {
				if len(pfile) > 0 {
					pfilei := []interface{}{}
					for pf := range pfile {
						pfilei = append(pfilei, pf)
					}
					params.SetFileParameter(pname, false, pfilei...)
					pfilei = nil
				}
			}
		}
	} else if err := r.ParseForm(); err == nil {
		if r.PostForm != nil {
			for pname, pvalue := range r.PostForm {
				params.SetParameter(pname, false, pvalue...)
			}
		} else if r.Form != nil {
			for pname, pvalue := range r.Form {
				params.SetParameter(pname, false, pvalue...)
			}
		}
	}
}

func (rtr *Router) serve(svr *Server, w http.ResponseWriter, r *http.Request) {
	defer func() {
		runtime.GC()
	}()

	rtr.rlock.RLock()

	var routePaths []string = []string{}

	params := NewParameters()
	params.SetParameter("next-request", false, r.URL.Path)

	defer func() {
		if params != nil {
			params.CleanupParameters()
			params = nil
		}
	}()

	setContentType := false

	loadParametersFromHttpRequest(params, r)

	var serveRouting func(string) error = func(pathToSplit string) (err error) {
		rtr.rlock.RLock()
		pathsplit := strings.Split(pathToSplit, "/")
		testPath := ""
		remainingPath := ""
		var routerFound *Route
		for nps := range pathsplit {
			if nps < len(pathsplit) {
				if testPath = strings.Join(pathsplit[0:len(pathsplit)-nps], "/"); testPath == "" {
					testPath = "/"
				} else if strings.HasSuffix(testPath, "/") {
					testPath = testPath[:len(testPath)-1]
				}

				remainingPath = strings.Join(pathsplit[len(pathsplit)-nps:], "/")
				if router, oktestpath := rtr.mappedRoutes[testPath]; oktestpath {
					routerFound = router
					break
				}
			}
		}
		rtr.rlock.RUnlock()
		if routerFound != nil {
			mimetype, extfound := FindMimeTypeByExt(filepath.Ext(remainingPath), ".txt")
			if !setContentType {
				setContentType = true
				w.Header().Set("CONTENT-TYPE", mimetype)
				w.WriteHeader(200)
			}
			var active *ActiveProcessor
			if isActiveExtension(extfound) {
				active = NewActiveProcessor(w)
				active.params = params
				active.canCleanupParams = false
			}
			defer func() {
				if active != nil {
					active.cleanupActiveProcessor()
					active = nil
				}
			}()
			routerFound.hndlefunc(svr, routerFound, routerFound.rootpath, remainingPath, mimetype, w, r, active)
		} else {
			w.Header().Set("CONTENT-TYPE", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte("no such route {" + testPath + "}"))
		}
		return err
	}
	lastRootPath := ""

	for {
		if params.ContainsParameter("next-request") && len(params.Parameter("next-request")) > 0 {
			if routePaths == nil {
				routePaths = []string{}
			}
			routePaths = append(routePaths, params.Parameter("next-request")...)
			params.SetParameter("next-request", true)
		} else if len(routePaths) == 0 {
			break
		} else if len(routePaths) > 0 {
			pathFound := routePaths[0]
			if len(routePaths) > 1 {
				routePaths = routePaths[1:]
			} else {
				routePaths = nil
			}
			if strings.Index(pathFound, "|") > -1 {
				for _, subPathFound := range strings.Split(pathFound, "|") {
					if strings.Index(subPathFound, "/") > -1 {
						lastRootPath = subPathFound[0 : strings.LastIndex(subPathFound, "/")+1]
						subPathFound = subPathFound[strings.LastIndex(subPathFound, "/")+1:]
					}
					if routePaths == nil {
						routePaths = []string{}
					}
					routePaths = append(routePaths, lastRootPath+subPathFound)
				}
			} else {
				if strings.Index(pathFound, "/") > -1 {
					lastRootPath = pathFound[0 : strings.LastIndex(pathFound, "/")+1]
					pathFound = pathFound[strings.LastIndex(pathFound, "/")+1:]
				}
				if err := serveRouting(lastRootPath + pathFound); err != nil {
					fmt.Printf("error: " + err.Error())
					break
				}
			}
		}
	}
}

var routes *Router = &Router{mappedRoutes: map[string]*Route{}, rlock: &sync.RWMutex{}}
