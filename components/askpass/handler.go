package askpass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dogmatiq/minibus"
)

// Handler is an [http.Handler] that publishes "askpass" requests to the message bus.
type Handler struct {
	init   sync.Once
	outbox chan<- any
	ready  chan struct{}

	pending sync.Map // request ID -> response channel
}

// Run pipes events received by the webhook handler to the message bus.
func (h *Handler) Run(ctx context.Context) (err error) {
	h.init.Do(func() {
		h.ready = make(chan struct{})
	})

	minibus.Subscribe[Response](ctx)
	minibus.Ready(ctx)

	h.outbox = minibus.Outbox(ctx)
	close(h.ready)

	for m := range minibus.Inbox(ctx) {
		res := m.(Response)
		if reply, ok := h.pending.Load(res.RequestID); ok {
			reply.(chan Response) <- res
		}
	}

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, ok := h.parseRequest(w, r)
	if !ok {
		return
	}

	response := make(chan Response, 1)
	h.pending.Store(req.RequestID, response)
	defer h.pending.Delete(req.RequestID)

	// wait for the outbox to become available
	h.init.Do(func() {
		h.ready = make(chan struct{})
	})

	select {
	case <-ctx.Done():
		h.writeContextError(ctx, w, "waiting for message bus initialization")
		return
	case <-h.ready:
	}

	// publish the request to the outbox
	select {
	case <-ctx.Done():
		h.writeContextError(ctx, w, "waiting to publish askpass request")
		return
	case h.outbox <- req:
	}

	// wait for the response
	select {
	case <-ctx.Done():
		h.writeContextError(ctx, w, "waiting for askpass response")
	case res := <-response:
		h.writeResponse(w, res)
	}
}

func (h *Handler) parseRequest(w http.ResponseWriter, r *http.Request) (Request, bool) {
	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return Request{}, false
	}

	repoURL, err := url.Parse(req.URL)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return Request{}, false
	}

	return Request{
		RequestID: req.ID,
		RepoURL:   repoURL,
		Field:     req.Field,
	}, true
}

func (h *Handler) writeResponse(w http.ResponseWriter, res Response) {
	data, err := json.Marshal(
		response{
			Value: res.Value,
		},
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) writeContextError(ctx context.Context, w http.ResponseWriter, message string) {
	code := http.StatusInternalServerError
	if ctx.Err() == context.DeadlineExceeded {
		code = http.StatusRequestTimeout
	}

	fmt.Printf(
		"%s: %s\n",
		message,
		ctx.Err(),
	)

	http.Error(
		w,
		fmt.Sprintf(
			"%s: %s",
			message,
			ctx.Err(),
		),
		code,
	)
}
