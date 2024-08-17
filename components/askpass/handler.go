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

	"github.com/dogmatiq/browser/model"
	"github.com/dogmatiq/minibus"
)

// Handler is an [http.Handler] that publishes "askpass" requests to the message bus.
type Handler struct {
	init   sync.Once
	outbox chan<- any
	ready  chan struct{}

	pending sync.Map // map[uuid.UUID]chan<- messages.CredentialResponse
}

// Run pipes events received by the webhook handler to the message bus.
func (h *Handler) Run(ctx context.Context) (err error) {
	h.init.Do(func() {
		h.ready = make(chan struct{})
	})
	h.outbox = minibus.Outbox(ctx)

	minibus.Subscribe[model.CredentialResponse](ctx)
	minibus.Ready(ctx)
	close(h.ready)

	for m := range minibus.Inbox(ctx) {
		switch m := m.(type) {
		case model.CredentialResponse:
			if reply, ok := h.pending.Load(m.RequestID); ok {
				reply.(chan model.CredentialResponse) <- m
			}
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

	h.init.Do(func() {
		h.ready = make(chan struct{})
	})

	// wait for the outbox to become available
	select {
	case <-ctx.Done():
		h.writeContextError(ctx, w, "waiting for message bus initialization")
		return
	case <-h.ready:
	}

	response := make(chan model.CredentialResponse, 1)
	h.pending.Store(req.RequestID, response)
	defer h.pending.Delete(req.RequestID)

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
		return
	case res := <-response:
		h.writeResponse(w, res)
	}
}

func (h *Handler) parseRequest(
	w http.ResponseWriter,
	r *http.Request,
) (model.CredentialRequest, bool) {
	var req apiRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return model.CredentialRequest{}, false
	}

	repoURL, err := url.Parse(req.URL)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return model.CredentialRequest{}, false
	}

	return model.CredentialRequest{
		RequestID:      req.ID,
		URL:            repoURL,
		CredentialType: req.Credential,
	}, true
}

func (h *Handler) writeResponse(
	w http.ResponseWriter,
	res model.CredentialResponse,
) {
	data, err := json.Marshal(
		apiResponse{
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
