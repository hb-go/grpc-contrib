package gateway

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/hb-go/grpc-contrib/client"
	"github.com/hb-go/grpc-contrib/gateway/codec"
	"github.com/hb-go/grpc-contrib/gateway/proto"
	"github.com/hb-go/grpc-contrib/gateway/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var valuesKeyRegexp = regexp.MustCompile("^(.*)\\[(.*)\\]$")

type Register struct {
	mux      runtime.ServeMuxDynamic
	registry registry.Registry

	sync.RWMutex
	// used to stop
	exit chan bool
	// indicate whether its running
	running bool
}

func (r *Register) run() {
	r.Lock()
	r.running = true
	r.Unlock()

	// reset watcher on exit
	defer func() {
		r.Lock()
		r.running = false
		r.Unlock()
	}()

	var a, b int
	for {
		// exit early if already dead
		if r.quit() {
			return
		}

		// jitter before starting
		j := rand.Int63n(100)
		time.Sleep(time.Duration(j) * time.Millisecond)

		// create new watcher
		w, err := r.registry.Watch()
		if err != nil {
			if r.quit() {
				return
			}

			dur := backoff(a)

			if a > 3 {
				grpclog.Warningf("gateway backing off: %d with error: %v", dur, err)
				a = 0
			}

			time.Sleep(dur)
			a++

			continue
		}

		// reset a
		a = 0

		// watch for events
		if err := r.watch(w); err != nil {
			if r.quit() {
				return
			}

			dur := backoff(b)

			if b > 3 {
				grpclog.Warningf("gateway backing off: %d with error: %v", dur, err)
				b = 0
			}

			time.Sleep(dur)
			b++

			continue
		}

		// reset b
		b = 0
	}
}

func (r *Register) watch(w registry.Watcher) error {
	// used to stop the watch
	stop := make(chan bool)

	// manage this loop
	go func() {
		defer w.Stop()

		select {
		// wait for exit
		case <-r.exit:
			return
		// we've been stopped
		case <-stop:
			return
		}
	}()

	for {
		res, err := w.Next()
		if err != nil {
			close(stop)
			return err
		}

		service := res.Service
		switch res.Action {
		case "create", "update":
			for _, m := range service.Methods {
				for _, route := range m.Routes {
					r.mux.Handle(
						route.Method,
						runtime.Pattern{},
						handler(service.Name, m.Name),
					)
				}
			}

		case "delete":
			for _, m := range service.Methods {
				for _, route := range m.Routes {
					r.mux.HandlerDeregister(
						route.Method,
						runtime.Pattern{},
					)
				}
			}
		}
	}
}

func (r *Register) quit() bool {
	select {
	case <-r.exit:
		return true
	default:
		return false
	}
}

func (r *Register) Start() {
	r.Lock()
	if !r.running {
		go r.run()
	}
	r.Unlock()
}

func (r *Register) Stop() {
	r.Lock()
	defer r.Unlock()

	select {
	case <-r.exit:
		return
	default:
		close(r.exit)
	}
}

func backoff(attempts int) time.Duration {
	if attempts == 0 {
		return time.Duration(0)
	}
	return time.Duration(math.Pow(10, float64(attempts))) * time.Millisecond
}

func handler(serviceName, method string) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		desc := grpc.ServiceDesc{ServiceName: serviceName}
		cc, closer, err := client.Client(
			&desc,
		)
		if err != nil {
			grpclog.Errorf("client error: %v", err)
		}
		defer closer.Close()

		payload := &simplejson.Json{}
		switch r.Method {
		case "PATCH", "POST", "PUT", "DELETE":
			// Body to JSON
			newReader, err := utilities.IOReaderFactory(r.Body)
			if err != nil {
				grpclog.Errorf("io reader factory error: %v", err)
			}
			payload, err = simplejson.NewFromReader(newReader())
			if err != nil {
				grpclog.Errorf("body parse error: %v", err)
			}
		}

		// Path params
		for key, val := range pathParams {
			fieldPath := strings.Split(key, ".")
			payload.SetPath(fieldPath, val)
		}

		// Query params
		// 已有数据不覆盖
		// 如果想要做到 grpc-gateway 那样根据 path 情况判断是否需要 query需要注册中心给出接口的 request 结构
		if err := r.ParseForm(); err != nil {
			grpclog.Errorf("request parse form error: %v", err)
		}
		for key, values := range r.Form {
			match := valuesKeyRegexp.FindStringSubmatch(key)
			if len(match) == 3 {
				key = match[1]
				values = append([]string{match[2]}, values...)
			}
			fieldPath := strings.Split(key, ".")

			if payload.GetPath(fieldPath...).Interface() == nil {
				payload.SetPath(fieldPath, values)
			}
		}

		data, err := payload.MarshalJSON()

		resp := &proto.Message{}
		req := proto.NewMessage(data)
		grpclog.Infof("req: %+v", req)

		method := fmt.Sprintf("/%s/%s", serviceName, method)
		err = cc.Invoke(context.TODO(), method, req, resp, grpc.CallContentSubtype(codec.CODEC_JSON))
	}
}

// New returns a new dynamic
func New(mux runtime.ServeMuxDynamic, r registry.Registry) *Register {
	return &Register{
		mux:      mux,
		registry: r,
		exit:     make(chan bool),
	}
}
