package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"knowledgegraph/internal/graph"
)

// ==================== 图谱 CRUD ====================

// handleListGraphs 获取图谱列表。
func (s *Server) handleListGraphs(w http.ResponseWriter, r *http.Request) {
	// 先查缓存
	if s.cache != nil && s.cache.IsEnabled() {
		if cached, err := s.cache.GetListJSON(r.Context()); err == nil && cached != "" && cached != "null" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Cache", "hit")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(cached))
			return
		}
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	list, err := s.store.ListGraphs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list graphs: "+err.Error())
		return
	}
	if list == nil {
		list = []*graph.GraphMeta{}
	}

	// 写缓存
	if s.cache != nil && s.cache.IsEnabled() {
		if data, err := json.Marshal(list); err == nil {
			_ = s.cache.SetListJSON(r.Context(), data)
		}
	}

	w.Header().Set("X-Cache", "miss")
	writeSuccess(w, list)
}

// handleGetGraph 获取完整图谱。
func (s *Server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	// 先查缓存
	if s.cache != nil && s.cache.IsEnabled() {
		if cached, err := s.cache.GetGraphJSON(r.Context(), graphCode); err == nil && cached != "" && cached != "null" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Cache", "hit")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(cached))
			return
		}
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	g, err := s.store.GetGraph(r.Context(), graphCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get graph: "+err.Error())
		return
	}
	if g == nil {
		writeError(w, http.StatusNotFound, "graph not found: "+graphCode)
		return
	}

	// 写缓存
	if s.cache != nil && s.cache.IsEnabled() {
		if data, err := json.Marshal(g); err == nil {
			_ = s.cache.SetGraphJSON(r.Context(), graphCode, data)
		}
	}

	w.Header().Set("X-Cache", "miss")
	writeSuccess(w, g)
}

// handleCreateGraph 创建图谱（全量批量写入）。
func (s *Server) handleCreateGraph(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	var payload graph.Payload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if payload.GraphCode == "" {
		writeError(w, http.StatusBadRequest, "graphCode is required")
		return
	}
	if !sanitizeGraphCode(payload.GraphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	// 检查是否已存在
	existingID, err := s.store.GetGraphIDByCode(r.Context(), payload.GraphCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check existing graph: "+err.Error())
		return
	}
	if existingID != 0 {
		writeError(w, http.StatusConflict, "graph already exists: "+payload.GraphCode)
		return
	}

	graphID, nodeCount, relCount, err := s.store.CreateGraph(r.Context(), &payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create graph: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), payload.GraphCode)

	writeJSON(w, http.StatusCreated, map[string]any{
		"success":        true,
		"graphCode":      payload.GraphCode,
		"version":        payload.Version,
		"graphId":        graphID,
		"nodeCount":      nodeCount,
		"relationCount":  relCount,
	})
}

// handleUpdateGraph 全量覆盖更新图谱。
func (s *Server) handleUpdateGraph(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	var payload graph.Payload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// 使用 URL 中的 graphCode 覆盖 payload 中的
	payload.GraphCode = graphCode

	graphID, nodeCount, relCount, err := s.store.UpdateGraph(r.Context(), graphCode, &payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update graph: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success":       true,
		"graphCode":     graphCode,
		"version":       payload.Version,
		"graphId":       graphID,
		"nodeCount":     nodeCount,
		"relationCount": relCount,
	})
}

// handleDeleteGraph 删除图谱。
func (s *Server) handleDeleteGraph(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	if err := s.store.DeleteGraph(r.Context(), graphCode); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete graph: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success":   true,
		"graphCode": graphCode,
	})
}

// handleValidateGraph 校验图谱完整性。
func (s *Server) handleValidateGraph(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	g, err := s.store.GetGraph(r.Context(), graphCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get graph: "+err.Error())
		return
	}
	if g == nil {
		writeError(w, http.StatusNotFound, "graph not found: "+graphCode)
		return
	}

	result := graph.ValidateGraph(g)
	writeSuccess(w, result)
}

// handleGraphStats 获取图谱统计。
func (s *Server) handleGraphStats(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	g, err := s.store.GetGraph(r.Context(), graphCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get graph: "+err.Error())
		return
	}
	if g == nil {
		writeError(w, http.StatusNotFound, "graph not found: "+graphCode)
		return
	}

	stats := graph.ComputeStats(g)
	writeSuccess(w, stats)
}

// ==================== 节点 CRUD ====================

// handleGetNode 获取节点（含上下文）。
func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	nodeID := pathValue(r, "nodeId")
	if !sanitizeGraphCode(graphCode) || nodeID == "" {
		writeError(w, http.StatusBadRequest, "invalid graphCode or nodeId")
		return
	}

	ctx, err := s.store.GetNode(r.Context(), graphCode, nodeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get node: "+err.Error())
		return
	}
	if ctx == nil {
		writeError(w, http.StatusNotFound, "node not found: "+nodeID)
		return
	}

	writeSuccess(w, ctx)
}

// handleCreateNode 添加节点。
func (s *Server) handleCreateNode(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	var node graph.Node
	if err := decodeJSONBody(r, &node); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if node.ID == "" || node.Label == "" {
		writeError(w, http.StatusBadRequest, "node id and label are required")
		return
	}
	if !graph.IsValidNodeType(node.Type) {
		writeError(w, http.StatusBadRequest, "invalid node type: "+node.Type)
		return
	}

	nodeID, err := s.store.CreateNode(r.Context(), graphCode, &node)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "graph not found: "+graphCode)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create node: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"nodeId":  nodeID,
	})
}

// handleUpdateNode 更新节点。
func (s *Server) handleUpdateNode(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	nodeID := pathValue(r, "nodeId")
	if !sanitizeGraphCode(graphCode) || nodeID == "" {
		writeError(w, http.StatusBadRequest, "invalid graphCode or nodeId")
		return
	}

	var node graph.Node
	if err := decodeJSONBody(r, &node); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// 使用 URL 中的 nodeID 覆盖
	node.ID = nodeID

	if node.Label == "" {
		writeError(w, http.StatusBadRequest, "node label is required")
		return
	}
	if !graph.IsValidNodeType(node.Type) {
		writeError(w, http.StatusBadRequest, "invalid node type: "+node.Type)
		return
	}

	resultID, err := s.store.UpdateNode(r.Context(), graphCode, nodeID, &node)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "node not found: "+nodeID)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update node: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success": true,
		"nodeId":  resultID,
	})
}

// handleDeleteNode 删除节点（含关联关系）。
func (s *Server) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	nodeID := pathValue(r, "nodeId")
	if !sanitizeGraphCode(graphCode) || nodeID == "" {
		writeError(w, http.StatusBadRequest, "invalid graphCode or nodeId")
		return
	}

	if err := s.store.DeleteNode(r.Context(), graphCode, nodeID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete node: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success": true,
		"nodeId":  nodeID,
	})
}

// ==================== 关系 CRUD ====================

// handleCreateRelation 添加关系。
func (s *Server) handleCreateRelation(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	if !sanitizeGraphCode(graphCode) {
		writeError(w, http.StatusBadRequest, "invalid graphCode")
		return
	}

	var rel graph.Relation
	if err := decodeJSONBody(r, &rel); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if rel.Source == "" || rel.Target == "" || rel.Type == "" {
		writeError(w, http.StatusBadRequest, "relation source, target and type are required")
		return
	}
	if !graph.IsValidRelationType(rel.Type) {
		writeError(w, http.StatusBadRequest, "invalid relation type: "+rel.Type)
		return
	}

	relationID, err := s.store.CreateRelation(r.Context(), graphCode, &rel)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "graph not found: "+graphCode)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create relation: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeJSON(w, http.StatusCreated, map[string]any{
		"success":     true,
		"relationId":  relationID,
	})
}

// handleUpdateRelation 更新关系。
func (s *Server) handleUpdateRelation(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	relationID := pathValue(r, "relationId")
	if !sanitizeGraphCode(graphCode) || relationID == "" {
		writeError(w, http.StatusBadRequest, "invalid graphCode or relationId")
		return
	}

	var rel graph.Relation
	if err := decodeJSONBody(r, &rel); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if rel.Source == "" || rel.Target == "" || rel.Type == "" {
		writeError(w, http.StatusBadRequest, "relation source, target and type are required")
		return
	}
	if !graph.IsValidRelationType(rel.Type) {
		writeError(w, http.StatusBadRequest, "invalid relation type: "+rel.Type)
		return
	}

	resultID, err := s.store.UpdateRelation(r.Context(), graphCode, relationID, &rel)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "relation not found: "+relationID)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update relation: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success":    true,
		"relationId": resultID,
	})
}

// handleDeleteRelation 删除关系。
func (s *Server) handleDeleteRelation(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	graphCode := pathValue(r, "graphCode")
	relationID := pathValue(r, "relationId")
	if !sanitizeGraphCode(graphCode) || relationID == "" {
		writeError(w, http.StatusBadRequest, "invalid graphCode or relationId")
		return
	}

	if err := s.store.DeleteRelation(r.Context(), graphCode, relationID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete relation: "+err.Error())
		return
	}

	// 失效缓存
	s.invalidateCache(r.Context(), graphCode)

	writeSuccess(w, map[string]any{
		"success":    true,
		"relationId": relationID,
	})
}

// ==================== 辅助函数 ====================

// decodeJSONBody 解码 JSON 请求体。
func decodeJSONBody(r *http.Request, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if len(body) == 0 {
		return errors.New("empty request body")
	}

	return json.Unmarshal(body, dst)
}

// invalidateCache 失效指定图谱的缓存和列表缓存。
func (s *Server) invalidateCache(ctx context.Context, graphCode string) {
	if s.cache == nil || !s.cache.IsEnabled() {
		return
	}
	_ = s.cache.DelGraph(ctx, graphCode)
	_ = s.cache.DelList(ctx)
}
