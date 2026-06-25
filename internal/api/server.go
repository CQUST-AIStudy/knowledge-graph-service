package api

import (
	"encoding/json"
	"log"
	"net/http"

	"knowledgegraph/internal/storage"
)

// Server 持有 HTTP 路由和依赖。
type Server struct {
	store        *storage.Store
	cache        *storage.CacheClient
	corsOrigins  []string
	allowAll     bool
}

// NewServer 创建 Server 实例。
func NewServer(store *storage.Store, cache *storage.CacheClient, corsOrigins []string) *Server {
	s := &Server{
		store:       store,
		cache:       cache,
		corsOrigins: corsOrigins,
	}
	// 规范化 CORS：只有 "*" 时允许所有
	for _, origin := range corsOrigins {
		if origin == "*" {
			s.allowAll = true
			break
		}
	}
	return s
}

// Routes 注册所有 HTTP 路由（使用 Go 1.22+ ServeMux 模式匹配）。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)

	// 图谱 CRUD
	mux.HandleFunc("GET /api/graphs", s.handleListGraphs)
	mux.HandleFunc("GET /api/graphs/{graphCode}", s.handleGetGraph)
	mux.HandleFunc("POST /api/graphs", s.handleCreateGraph)
	mux.HandleFunc("PUT /api/graphs/{graphCode}", s.handleUpdateGraph)
	mux.HandleFunc("DELETE /api/graphs/{graphCode}", s.handleDeleteGraph)
	mux.HandleFunc("GET /api/graphs/{graphCode}/validate", s.handleValidateGraph)
	mux.HandleFunc("GET /api/graphs/{graphCode}/stats", s.handleGraphStats)

	// 节点 CRUD
	mux.HandleFunc("GET /api/graphs/{graphCode}/nodes/{nodeId}", s.handleGetNode)
	mux.HandleFunc("POST /api/graphs/{graphCode}/nodes", s.handleCreateNode)
	mux.HandleFunc("PUT /api/graphs/{graphCode}/nodes/{nodeId}", s.handleUpdateNode)
	mux.HandleFunc("DELETE /api/graphs/{graphCode}/nodes/{nodeId}", s.handleDeleteNode)

	// 关系 CRUD
	mux.HandleFunc("POST /api/graphs/{graphCode}/relations", s.handleCreateRelation)
	mux.HandleFunc("PUT /api/graphs/{graphCode}/relations/{relationId}", s.handleUpdateRelation)
	mux.HandleFunc("DELETE /api/graphs/{graphCode}/relations/{relationId}", s.handleDeleteRelation)

	return s.withCORS(mux)
}

// ---- CORS 中间件 ----

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 预检请求
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", s.originFor(r))
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		origin := s.originFor(r)
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if !s.allowAll {
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		next.ServeHTTP(w, r)
	})
}

func (s *Server) originFor(r *http.Request) string {
	if s.allowAll {
		return "*"
	}
	requestOrigin := r.Header.Get("Origin")
	for _, allowed := range s.corsOrigins {
		if allowed == requestOrigin {
			return requestOrigin
		}
	}
	return ""
}

// ---- 响应辅助函数 ----

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("api: write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error":   message,
	})
}

func writeSuccess(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

// pathValue 安全读取路径参数。
func pathValue(r *http.Request, key string) string {
	return r.PathValue(key)
}

// sanitizeGraphCode 校验 graphCode 合法性。
func sanitizeGraphCode(code string) bool {
	if code == "" || len(code) > 128 {
		return false
	}
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
