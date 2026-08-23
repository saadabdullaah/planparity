package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (server *Server) handleEvents(writer http.ResponseWriter, request *http.Request) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		writeAPIError(writer, http.StatusInternalServerError, "STREAM_UNSUPPORTED", "Event streaming is unavailable.")
		return
	}
	afterID := int64(0)
	if value := request.Header.Get("Last-Event-ID"); value != "" {
		afterID, _ = strconv.ParseInt(value, 10, 64)
	}
	events, cancel, exists := server.jobs.Subscribe(request.PathValue("id"), afterID)
	if !exists {
		writeAPIError(writer, http.StatusNotFound, "JOB_NOT_FOUND", "Job was not found.")
		return
	}
	defer cancel()
	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	flusher.Flush()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case event, open := <-events:
			if !open {
				return
			}
			payload, _ := json.Marshal(event)
			fmt.Fprintf(writer, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, payload)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprint(writer, ": keep-alive\n\n")
			flusher.Flush()
		case <-request.Context().Done():
			return
		}
	}
}
