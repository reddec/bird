package gazer

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/reddec/bird"
)

const defaultLimit = 100

// Nest where new birds can be obtained
type Nest func(params url.Values) (bird.Bird, error)

type birdFace struct {
	Name     string        `json:"name"`
	Flying   bool          `json:"flying"`
	Interval time.Duration `json:"interval"`
}

type gazer struct {
	flock *bird.Flock
	nest  Nest
}

func (gz *gazer) listAllBirds(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// pagination
	var limit int64 = defaultLimit
	var offset int64
	var err error
	if r.FormValue("limit") != "" {
		limit, err = strconv.ParseInt(r.FormValue("limit"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.FormValue("offset") != "" {
		offset, err = strconv.ParseInt(r.FormValue("offset"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// filter by bird name
	filter := r.FormValue("name")
	var birds []*bird.SmartBird
	if filter != "" {
		birds = gz.flock.Select(filter)
	} else {
		birds = gz.flock.Select()
	}
	// check bounds
	count := int64(len(birds))
	if offset >= count {
		offset = count
	}
	if offset+limit >= count {
		limit = count - offset
	}

	// prepare answer
	var faces []birdFace
	for _, bird := range birds[offset : offset+limit] {
		faces = append(faces, birdFace{Name: bird.Name(), Interval: bird.Interval(), Flying: bird.Flying()})
	}
	data, err := json.MarshalIndent(faces, "", "    ")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// add some REST sugar for correct response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (gz *gazer) controlBirds(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var names []string
	name := r.FormValue("name")
	if name != "" {
		names = append(names, name)
	}
	action := r.FormValue("action")
	if action == "" {
		http.Error(w, "no action specified", http.StatusBadRequest)
		return
	}
	switch action {
	case "land":
		gz.flock.Land(names...)
	case "raise":
		gz.flock.Raise(names...)
	default:
		http.Error(w, "unknown action specified: "+action, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (gz *gazer) killBirds(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var names []string
	name := r.FormValue("name")
	if name != "" {
		names = append(names, name)
	}
	gz.flock.Exclude(true, name)
	w.WriteHeader(http.StatusOK)
}

func (gz *gazer) includeBird(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if gz.nest == nil {
		http.Error(w, "Nest not provided", http.StatusNotAcceptable)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "'name' param required", http.StatusBadRequest)
		return
	}
	interval, err := time.ParseDuration(r.FormValue("interval"))
	if err != nil {
		http.Error(w, "'interval' param parsing: "+err.Error(), http.StatusBadRequest)
		return
	}
	birdFunc, err := gz.nest(r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raise := r.FormValue("raise") != ""
	smartBird := bird.NewSmartBird(birdFunc, interval, name)
	if raise {
		smartBird.Start()
	}
	gz.flock.Include(smartBird)
	// Write answer
	data, err := json.MarshalIndent(birdFace{
		Name:     smartBird.Name(),
		Interval: smartBird.Interval(),
		Flying:   smartBird.Flying()}, "", "    ")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// NewGazer is a constructor of JSON API.
// It can create new bird by nest, remove, land/raise and select birds from flock
// Methods:
//
// GET - select birds (optional param 'name' as filter)
//
// PUT - control bird. Required param 'action' with value 'raise' or 'land'
//
// POST - add new bird by nest.  Required param 'interval' (duration) and 'name'. If required autoraise - use param 'raise' (any non-empty value)
//
// DELETE - land and kill bird
func NewGazer(flock *bird.Flock, nest Nest) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	gz := &gazer{flock, nest}
	router.Methods("GET").HandlerFunc(gz.listAllBirds)
	router.Methods("PUT").HandlerFunc(gz.controlBirds)
	router.Methods("POST").HandlerFunc(gz.includeBird)
	router.Methods("DELETE").HandlerFunc(gz.killBirds)
	return router
}
