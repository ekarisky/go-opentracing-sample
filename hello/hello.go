package hello

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
)

type HelloModule struct {
	// TODO: struct your module here
}

func NewHelloModule() *HelloModule {
	// TODO: do some config & init here
	return &HelloModule{}
}

// each handler can return the data and error, and ServeHTTP can chose how to convert this
type HandlerFunc func(rw http.ResponseWriter, r *http.Request) (interface{}, []string, error)

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Tkpd-UserId,Authorization,Origin")

	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	span, ctx := opentracing.StartSpanFromContext(r.Context(), r.URL.Path)
	defer span.Finish()

	ctx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	r = r.WithContext(ctx)

	if userid := r.Header.Get("Tkpd-UserId"); userid != "" {
		span.LogFields(
			tlog.String("user_id", userid),
			tlog.String("ip", r.Header.Get("X-Forwarded-For")))
	}

	start := time.Now()

	var data interface{}
	var err error

	errStatus := http.StatusInternalServerError

	data, msg, err := fn(w, r)

	response := Response{}
	response.Base.Status = "OK"
	response.Base.ServerProcessTime = time.Since(start).String()

	var buf []byte

	w.Header().Set("Content-Type", "application/json")

	if msg != nil {
		response.StatusMessage = msg
	}

	if err == nil {
		response.Data = data
		if buf, err = json.Marshal(response); err == nil {
			w.Write(buf)
			return
		}
	}

	if err != nil {
		ext.Error.Set(span, true)
		response.Base.ErrorMessage = []string{
			err.Error(),
		}
		buf, _ := json.Marshal(response.Base)
		log.Println("handler error", err.Error(), string(buf[:]))
		w.WriteHeader(errStatus)
		w.Write(buf)
		return
	}
}

func (h *HelloModule) InitHandlers() {
	http.Handle("/ping", HandlerFunc(h.Ping))
}

// status check url for the uberapp
func (h *HelloModule) Ping(w http.ResponseWriter, r *http.Request) (interface{}, []string, error) {
	// TODO: check connectivity with db
	getDataFromDB(r.Context())

	// TODO: check connectivity with redis
	getDataFromRedis(r.Context())

	// TODO: check connectivity with api
	go getDataFromAPI(r.Context())

	// TODO: check connectivity with some stuff

	return "PONG", nil, nil
}

func getDataFromDB(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getDataFromDB")
	defer span.Finish()

	time.Sleep(time.Millisecond * 500)
}

func getDataFromRedis(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getDataFromRedis")
	defer span.Finish()

	time.Sleep(time.Millisecond * 100)
}

func getDataFromAPI(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getDataFromAPI")
	defer span.Finish()

	time.Sleep(time.Millisecond * 1200)
}
