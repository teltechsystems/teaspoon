package router

import (
	"fmt"
	"github.com/teltechsystems/teaspoon"
	"sync"
)

func NotFound(w teaspoon.ResponseWriter, r *teaspoon.Request) {
	fmt.Println("NOT FOUND!!")
}

type Router struct {
	handlers map[int]teaspoon.Handler
	mu       sync.RWMutex
	notFound teaspoon.Handler
}

func (router *Router) Handle(resource int, handler teaspoon.Handler) {
	router.mu.Lock()
	defer router.mu.Unlock()

	if router.handlers == nil {
		router.handlers = make(map[int]teaspoon.Handler)
	}

	router.handlers[resource] = handler
}

func (router *Router) HandleFunc(resource int, handlerFunc func(teaspoon.ResponseWriter, *teaspoon.Request)) {
	router.Handle(resource, teaspoon.HandlerFunc(handlerFunc))
}

func (router *Router) ServeTSP(w teaspoon.ResponseWriter, r *teaspoon.Request) {
	router.mu.RLock()
	defer router.mu.RUnlock()

	handler, ok := router.handlers[r.Resource]
	if !ok {
		handler = router.notFound
	}

	handler.ServeTSP(w, r)
}

func NewRouter(notFound teaspoon.Handler) *Router {
	if notFound == nil {
		notFound = teaspoon.HandlerFunc(NotFound)
	}

	return &Router{notFound: notFound}
}
