package api

import (
	"net/http"
)

// handleHealth 健康检查（始终 200，报告 DB/Redis 状态）。
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status": "ok",
		"service": "knowledge-graph-service",
	}

	// MySQL 状态
	mysqlStatus := "disabled"
	if s.store != nil {
		if err := s.store.Ping(r.Context()); err != nil {
			mysqlStatus = "error"
		} else {
			mysqlStatus = "ok"
		}
	}
	resp["mysql"] = mysqlStatus

	// Redis 状态
	redisStatus := "disabled"
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.Ping(r.Context()); err != nil {
			redisStatus = "error"
		} else {
			redisStatus = "ok"
		}
	}
	resp["redis"] = redisStatus

	writeJSON(w, http.StatusOK, resp)
}

// handleReady 就绪探测（DB 可用才 200，否则 503）。
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status": "ok",
		"service": "knowledge-graph-service",
	}

	status := http.StatusOK

	// MySQL 必须可用
	mysqlStatus := "ok"
	if s.store == nil {
		mysqlStatus = "disabled"
		status = http.StatusServiceUnavailable
	} else if err := s.store.Ping(r.Context()); err != nil {
		mysqlStatus = "error"
		status = http.StatusServiceUnavailable
	}
	resp["mysql"] = mysqlStatus

	// Redis 可选
	redisStatus := "disabled"
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.Ping(r.Context()); err != nil {
			redisStatus = "error"
		} else {
			redisStatus = "ok"
		}
	}
	resp["redis"] = redisStatus

	if status != http.StatusOK {
		resp["status"] = "not ready"
	}

	writeJSON(w, status, resp)
}
