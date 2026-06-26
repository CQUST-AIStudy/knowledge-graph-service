package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"knowledgegraph/internal/graph"
)

// Store 封装 MySQL 数据访问。
type Store struct {
	db                     *sql.DB
	progressCompletionSQL  string
	progressCompletionArgs []string
}

const defaultGraphCode = "data-structure-knowledge-graph"

const defaultProgressCompletionSQL = `
SELECT DISTINCT
	CASE
		WHEN ao.source_system = 'LEGACY_TAP'
			AND ao.source_offering_key LIKE 'LEGACY_EXPERIMENT_OFFERING:%'
		THEN CAST(CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(ao.source_offering_key, ':CLASS:', 1), ':', -1) AS SIGNED) AS CHAR)
		ELSE CAST(sa.offering_id AS CHAR)
	END AS exercise_key
FROM student_assignment sa
JOIN student_profile sp ON sp.id = sa.student_id
JOIN assignment_offering ao ON ao.id = sa.offering_id
WHERE (
	sp.student_no = ?
	OR CAST(sp.id AS CHAR) = ?
	OR CAST(sp.user_id AS CHAR) = ?
)
AND (
	LOWER(COALESCE(sa.submission_status, '')) IN ('graded', 'submitted', 'closed')
	OR COALESCE(sa.completion_evidence, 'NONE') IN ('TRANSCRIPT_SCORE', 'ANSWER_SHEET', 'SCORED_CODE')
)`

var defaultProgressCompletionArgs = []string{"userId", "userId", "userId"}

// NewMySQLStore 创建并返回一个 MySQL 存储实例。
func NewMySQLStore(dsn string) (*Store, error) {
	if dsn == "" {
		return nil, errors.New("mysql: DSN is empty")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &Store{db: db}, nil
}

// ConfigureProgressCompletionQuery 配置用户进度的只读完成记录查询。
func (s *Store) ConfigureProgressCompletionQuery(query string, args []string) {
	if s == nil {
		return
	}
	s.progressCompletionSQL = strings.TrimSpace(query)
	s.progressCompletionArgs = normalizeProgressArgTokens(args)
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Ping 检查数据库连通性。
func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("mysql: store is nil")
	}
	return s.db.PingContext(ctx)
}

// EnsureSchema 确保必要的表存在。
func (s *Store) EnsureSchema(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("mysql: store is nil")
	}
	tables := []string{
		`CREATE TABLE IF NOT EXISTS kg_graph (
			id             BIGINT AUTO_INCREMENT PRIMARY KEY,
			graph_code     VARCHAR(128) NOT NULL,
			version        VARCHAR(32)  NOT NULL DEFAULT '1.0.0',
			source_json    JSON,
			metadata_json  JSON,
			course_node_id VARCHAR(128),
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_graph_code (graph_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS kg_node (
			id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
			graph_id            BIGINT NOT NULL,
			node_id             VARCHAR(128) NOT NULL,
			label               VARCHAR(256) NOT NULL,
			type                VARCHAR(32) NOT NULL,
			chapter_id          VARCHAR(128),
			summary             TEXT,
			properties_json     JSON,
			prerequisites_json  JSON,
			related_json        JSON,
			applies_to_json     JSON,
			targets_json        JSON,
			sort_order          INT DEFAULT 0,
			created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_graph_node (graph_id, node_id),
			INDEX idx_graph (graph_id),
			INDEX idx_type (graph_id, type),
			INDEX idx_chapter (graph_id, chapter_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS kg_relation (
			id              BIGINT AUTO_INCREMENT PRIMARY KEY,
			graph_id        BIGINT NOT NULL,
			relation_id     VARCHAR(256) NOT NULL,
			source          VARCHAR(128) NOT NULL,
			target          VARCHAR(128) NOT NULL,
			type            VARCHAR(32) NOT NULL,
			properties_json JSON,
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_graph_relation (graph_id, relation_id),
			INDEX idx_graph (graph_id),
			INDEX idx_source (graph_id, source),
			INDEX idx_target (graph_id, target)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, ddl := range tables {
		if _, err := s.db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("mysql: ensure schema: %w", err)
		}
	}
	return nil
}

// EnsureDefaultGraphSeed 在默认图谱缺失或为空时导入内置种子数据。
func (s *Store) EnsureDefaultGraphSeed(ctx context.Context, script string) (bool, error) {
	if s == nil || s.db == nil {
		return false, errors.New("mysql: store is nil")
	}
	if strings.TrimSpace(script) == "" {
		return false, errors.New("mysql: seed script is empty")
	}

	graphID, err := s.GetGraphIDByCode(ctx, defaultGraphCode)
	if err != nil {
		return false, fmt.Errorf("mysql: check default graph: %w", err)
	}
	if graphID != 0 {
		var nodeCount int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM kg_node WHERE graph_id = ?`, graphID).Scan(&nodeCount); err != nil {
			return false, fmt.Errorf("mysql: count default graph nodes: %w", err)
		}
		if nodeCount > 0 {
			return false, nil
		}
	}

	if err := s.execSQLScript(ctx, script); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) execSQLScript(ctx context.Context, script string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: begin seed tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	statements := splitSQLStatements(stripSQLLineComments(script))
	for i, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("mysql: execute seed statement %d: %w", i+1, err)
		}
	}
	return tx.Commit()
}

func stripSQLLineComments(script string) string {
	var b strings.Builder
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func splitSQLStatements(script string) []string {
	parts := strings.Split(script, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}

// ==================== 图谱 CRUD ====================

// GetGraphIDByCode 根据 graphCode 获取图谱内部 ID。
func (s *Store) GetGraphIDByCode(ctx context.Context, graphCode string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM kg_graph WHERE graph_code = ?`, graphCode,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// ListGraphs 返回所有图谱的摘要信息。
func (s *Store) ListGraphs(ctx context.Context) ([]*graph.GraphMeta, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.graph_code, g.version, g.updated_at,
		        (SELECT COUNT(*) FROM kg_node n WHERE n.graph_id = g.id) AS node_count,
		        (SELECT COUNT(*) FROM kg_relation r WHERE r.graph_id = g.id) AS relation_count,
		        JSON_UNQUOTE(JSON_EXTRACT(g.metadata_json, '$.title')) AS title
		 FROM kg_graph g
		 ORDER BY g.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("mysql: list graphs: %w", err)
	}
	defer rows.Close()

	var result []*graph.GraphMeta
	for rows.Next() {
		var m graph.GraphMeta
		var title sql.NullString
		if err := rows.Scan(&m.GraphCode, &m.Version, &m.UpdatedAt, &m.NodeCount, &m.RelationCount, &title); err != nil {
			return nil, err
		}
		m.Title = title.String
		if m.Title == "" {
			m.Title = m.GraphCode
		}
		result = append(result, &m)
	}
	return result, rows.Err()
}

// GetGraph 获取完整图谱（元数据 + 课程节点 + 所有节点 + 所有关系）。
func (s *Store) GetGraph(ctx context.Context, graphCode string) (*graph.Graph, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return nil, fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return nil, nil // 不存在
	}

	// 读取图谱元数据
	var (
		metadataJSON []byte
		courseNodeID sql.NullString
	)
	err = s.db.QueryRowContext(ctx,
		`SELECT metadata_json, course_node_id FROM kg_graph WHERE id = ?`, graphID,
	).Scan(&metadataJSON, &courseNodeID)
	if err != nil {
		return nil, fmt.Errorf("mysql: get graph meta: %w", err)
	}

	// 读取所有节点
	nodeRows, err := s.queryNodes(ctx, graphID)
	if err != nil {
		return nil, err
	}

	// 读取所有关系
	relationRows, err := s.queryRelations(ctx, graphID)
	if err != nil {
		return nil, err
	}

	// 组装 Graph
	g := &graph.Graph{
		Metadata:  graph.ParseMapRaw(metadataJSON),
		Nodes:     make([]*graph.Node, 0, len(nodeRows)),
		Relations: make([]*graph.Relation, 0, len(relationRows)),
	}

	// 找到课程节点
	for _, nr := range nodeRows {
		node := graph.RowToNode(nr)
		if nr.NodeID == courseNodeID.String || nr.Type == "course" {
			g.Course = node
		}
		g.Nodes = append(g.Nodes, node)
	}

	for _, rr := range relationRows {
		g.Relations = append(g.Relations, graph.RowToRelation(rr))
	}

	return g, nil
}

// GetGraphProgress 获取指定用户在图谱上的学习进度。
func (s *Store) GetGraphProgress(ctx context.Context, graphCode, userID string) (*graph.GraphProgress, error) {
	g, err := s.GetGraph(ctx, graphCode)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	keys, source, err := s.QueryCompletedExerciseKeys(ctx, graphCode, userID)
	if err != nil {
		return nil, err
	}
	return graph.ComputeProgress(graphCode, g, userID, keys, source), nil
}

// GetGraphWithProgress 获取图谱并在节点上附加指定用户的学习进度。
func (s *Store) GetGraphWithProgress(ctx context.Context, graphCode, userID string) (*graph.Graph, error) {
	g, err := s.GetGraph(ctx, graphCode)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	keys, source, err := s.QueryCompletedExerciseKeys(ctx, graphCode, userID)
	if err != nil {
		return nil, err
	}
	progress := graph.ComputeProgress(graphCode, g, userID, keys, source)
	graph.ApplyProgressToGraph(g, progress)
	return g, nil
}

// QueryCompletedExerciseKeys 读取用户已完成练习标识；返回值可匹配 exercise node_id 或 properties.experimentId。
func (s *Store) QueryCompletedExerciseKeys(ctx context.Context, graphCode, userID string) ([]string, string, error) {
	if s == nil || s.db == nil {
		return nil, "", errors.New("mysql: store is nil")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, "", errors.New("mysql: userId is required")
	}
	query, argTokens, source := s.progressCompletionQuery()
	args, err := buildProgressQueryArgs(argTokens, graphCode, userID)
	if err != nil {
		return nil, source, err
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, source, fmt.Errorf("mysql: query completed exercise keys: %w", err)
	}
	defer rows.Close()

	keys := make([]string, 0)
	seen := map[string]struct{}{}
	for rows.Next() {
		var key sql.NullString
		if err := rows.Scan(&key); err != nil {
			return nil, source, fmt.Errorf("mysql: scan completed exercise key: %w", err)
		}
		value := strings.TrimSpace(key.String)
		if !key.Valid || value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		keys = append(keys, value)
	}
	if err := rows.Err(); err != nil {
		return nil, source, fmt.Errorf("mysql: iterate completed exercise keys: %w", err)
	}
	return keys, source, nil
}

// CreateGraph 创建图谱并批量写入节点和关系（事务）。
func (s *Store) CreateGraph(ctx context.Context, p *graph.Payload) (int64, int, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. 插入图谱元数据
	sourceJSON, _ := json.Marshal(p.Source)
	metadataJSON, _ := json.Marshal(p.Metadata)
	courseNodeID := ""
	if p.Course != nil {
		courseNodeID = p.Course.ID
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO kg_graph (graph_code, version, source_json, metadata_json, course_node_id)
		 VALUES (?, ?, ?, ?, ?)`,
		p.GraphCode, p.Version, sourceJSON, metadataJSON, courseNodeID,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: insert graph: %w", err)
	}
	graphID, _ := result.LastInsertId()

	// 2. 批量插入节点（含 course 节点）
	allNodes := p.Nodes
	if p.Course != nil {
		// course 节点单独处理，确保在列表中
		alreadyExists := false
		for _, n := range allNodes {
			if n.ID == p.Course.ID {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			allNodes = append([]*graph.Node{p.Course}, allNodes...)
		}
	}

	nodeCount, err := s.batchInsertNodes(ctx, tx, graphID, allNodes)
	if err != nil {
		return 0, 0, 0, err
	}

	// 3. 批量插入关系
	relationCount, err := s.batchInsertRelations(ctx, tx, graphID, p.Relations)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: commit: %w", err)
	}

	return graphID, nodeCount, relationCount, nil
}

// UpdateGraph 全量覆盖更新图谱（删除旧节点/关系，写入新的）。
func (s *Store) UpdateGraph(ctx context.Context, graphCode string, p *graph.Payload) (int64, int, int, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		// 不存在则创建
		return s.CreateGraph(ctx, p)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. 更新元数据
	sourceJSON, _ := json.Marshal(p.Source)
	metadataJSON, _ := json.Marshal(p.Metadata)
	courseNodeID := ""
	if p.Course != nil {
		courseNodeID = p.Course.ID
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE kg_graph SET version = ?, source_json = ?, metadata_json = ?, course_node_id = ? WHERE id = ?`,
		p.Version, sourceJSON, metadataJSON, courseNodeID, graphID,
	); err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: update graph: %w", err)
	}

	// 2. 删除旧节点和关系
	if _, err := tx.ExecContext(ctx, `DELETE FROM kg_relation WHERE graph_id = ?`, graphID); err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: delete old relations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM kg_node WHERE graph_id = ?`, graphID); err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: delete old nodes: %w", err)
	}

	// 3. 批量插入新节点
	allNodes := p.Nodes
	if p.Course != nil {
		alreadyExists := false
		for _, n := range allNodes {
			if n.ID == p.Course.ID {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			allNodes = append([]*graph.Node{p.Course}, allNodes...)
		}
	}
	nodeCount, err := s.batchInsertNodes(ctx, tx, graphID, allNodes)
	if err != nil {
		return 0, 0, 0, err
	}

	// 4. 批量插入新关系
	relationCount, err := s.batchInsertRelations(ctx, tx, graphID, p.Relations)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("mysql: commit: %w", err)
	}

	return graphID, nodeCount, relationCount, nil
}

// DeleteGraph 删除图谱及其所有节点和关系。
func (s *Store) DeleteGraph(ctx context.Context, graphCode string) error {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return nil // 不存在视为成功
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM kg_relation WHERE graph_id = ?`, graphID); err != nil {
		return fmt.Errorf("mysql: delete relations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM kg_node WHERE graph_id = ?`, graphID); err != nil {
		return fmt.Errorf("mysql: delete nodes: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM kg_graph WHERE id = ?`, graphID); err != nil {
		return fmt.Errorf("mysql: delete graph: %w", err)
	}

	return tx.Commit()
}

// ==================== 节点 CRUD ====================

// GetNode 获取单个节点（含上下文）。
func (s *Store) GetNode(ctx context.Context, graphCode, nodeID string) (*graph.NodeContext, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return nil, fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return nil, nil
	}

	// 读取目标节点
	nodeRow, err := s.queryNodeByID(ctx, graphID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("mysql: get node: %w", err)
	}
	if nodeRow == nil {
		return nil, nil
	}
	node := graph.RowToNode(nodeRow)

	// 读取所有节点构建 nodeMap（用于上下文查找）
	nodeRows, err := s.queryNodes(ctx, graphID)
	if err != nil {
		return nil, err
	}
	nodeMap := make(map[string]*graph.Node, len(nodeRows))
	for _, nr := range nodeRows {
		nodeMap[nr.NodeID] = graph.RowToNode(nr)
	}

	// 读取所有关系
	relationRows, err := s.queryRelations(ctx, graphID)
	if err != nil {
		return nil, err
	}

	// 构建上下文
	ctx_ := &graph.NodeContext{
		Node:          node,
		AncestorChain: []*graph.Node{},
		Prerequisites: []*graph.Node{},
		NextNodes:     []*graph.Node{},
		RelatedNodes:  []*graph.Node{},
		Exercises:     []*graph.Node{},
	}

	// 祖先链
	if node.Type != "course" {
		if node.Type == "chapter" {
			// 找课程节点
			for _, nr := range nodeRows {
				if nr.Type == "course" {
					ctx_.AncestorChain = append(ctx_.AncestorChain, graph.RowToNode(nr))
					break
				}
			}
		}
		ctx_.AncestorChain = append(ctx_.AncestorChain, node)
		if node.ChapterID != nil {
			if ch, ok := nodeMap[*node.ChapterID]; ok {
				// 插入到倒数第二个位置（课程→章节→节点）
				ctx_.AncestorChain = append([]*graph.Node{ch}, ctx_.AncestorChain...)
			}
		}
	}

	// 遍历关系构建上下文
	for _, rr := range relationRows {
		switch rr.Type {
		case "PREREQUISITE":
			if rr.Target == nodeID {
				if n, ok := nodeMap[rr.Source]; ok {
					ctx_.Prerequisites = append(ctx_.Prerequisites, n)
				}
			}
			if rr.Source == nodeID {
				if n, ok := nodeMap[rr.Target]; ok {
					ctx_.NextNodes = append(ctx_.NextNodes, n)
				}
			}
		case "RELATED_TO":
			if rr.Target == nodeID {
				if n, ok := nodeMap[rr.Source]; ok {
					ctx_.RelatedNodes = append(ctx_.RelatedNodes, n)
				}
			}
			if rr.Source == nodeID {
				if n, ok := nodeMap[rr.Target]; ok {
					ctx_.RelatedNodes = append(ctx_.RelatedNodes, n)
				}
			}
		case "APPLIES_TO":
			if rr.Source == nodeID {
				if n, ok := nodeMap[rr.Target]; ok {
					ctx_.RelatedNodes = append(ctx_.RelatedNodes, n)
				}
			}
		case "TESTED_BY":
			if rr.Source == nodeID {
				if n, ok := nodeMap[rr.Target]; ok {
					ctx_.Exercises = append(ctx_.Exercises, n)
				}
			}
		}
	}

	return ctx_, nil
}

// CreateNode 添加节点。
func (s *Store) CreateNode(ctx context.Context, graphCode string, node *graph.Node) (string, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return "", fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return "", errors.New("mysql: graph not found")
	}

	propsJSON, _ := json.Marshal(node.Properties)
	prereqJSON, _ := json.Marshal(node.Prerequisites)
	relatedJSON, _ := json.Marshal(node.Related)
	appliesJSON, _ := json.Marshal(node.AppliesTo)
	targetsJSON, _ := json.Marshal(node.Targets)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, prerequisites_json, related_json, applies_to_json, targets_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		graphID, node.ID, node.Label, node.Type, node.ChapterID, node.Summary,
		propsJSON, prereqJSON, relatedJSON, appliesJSON, targetsJSON,
	)
	if err != nil {
		return "", fmt.Errorf("mysql: insert node: %w", err)
	}
	return node.ID, nil
}

// UpdateNode 更新节点。
func (s *Store) UpdateNode(ctx context.Context, graphCode, nodeID string, node *graph.Node) (string, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return "", fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return "", errors.New("mysql: graph not found")
	}

	propsJSON, _ := json.Marshal(node.Properties)
	prereqJSON, _ := json.Marshal(node.Prerequisites)
	relatedJSON, _ := json.Marshal(node.Related)
	appliesJSON, _ := json.Marshal(node.AppliesTo)
	targetsJSON, _ := json.Marshal(node.Targets)

	result, err := s.db.ExecContext(ctx,
		`UPDATE kg_node SET label = ?, type = ?, chapter_id = ?, summary = ?, properties_json = ?, prerequisites_json = ?, related_json = ?, applies_to_json = ?, targets_json = ?
		 WHERE graph_id = ? AND node_id = ?`,
		node.Label, node.Type, node.ChapterID, node.Summary,
		propsJSON, prereqJSON, relatedJSON, appliesJSON, targetsJSON,
		graphID, nodeID,
	)
	if err != nil {
		return "", fmt.Errorf("mysql: update node: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "", errors.New("mysql: node not found")
	}
	return nodeID, nil
}

// DeleteNode 删除节点及其关联关系。
func (s *Store) DeleteNode(ctx context.Context, graphCode, nodeID string) error {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 删除涉及该节点的关系
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM kg_relation WHERE graph_id = ? AND (source = ? OR target = ?)`,
		graphID, nodeID, nodeID,
	); err != nil {
		return fmt.Errorf("mysql: delete relations for node: %w", err)
	}

	// 删除节点
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM kg_node WHERE graph_id = ? AND node_id = ?`,
		graphID, nodeID,
	); err != nil {
		return fmt.Errorf("mysql: delete node: %w", err)
	}

	return tx.Commit()
}

// ==================== 关系 CRUD ====================

// CreateRelation 添加关系。
func (s *Store) CreateRelation(ctx context.Context, graphCode string, rel *graph.Relation) (string, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return "", fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return "", errors.New("mysql: graph not found")
	}

	relID := rel.ID
	if relID == "" {
		relID = graph.MakeRelationID(rel.Source, rel.Target, rel.Type)
	}
	propsJSON, _ := json.Marshal(rel.Properties)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		graphID, relID, rel.Source, rel.Target, rel.Type, propsJSON,
	)
	if err != nil {
		return "", fmt.Errorf("mysql: insert relation: %w", err)
	}
	return relID, nil
}

// UpdateRelation 更新关系。
func (s *Store) UpdateRelation(ctx context.Context, graphCode, relationID string, rel *graph.Relation) (string, error) {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return "", fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return "", errors.New("mysql: graph not found")
	}

	propsJSON, _ := json.Marshal(rel.Properties)

	result, err := s.db.ExecContext(ctx,
		`UPDATE kg_relation SET source = ?, target = ?, type = ?, properties_json = ?
		 WHERE graph_id = ? AND relation_id = ?`,
		rel.Source, rel.Target, rel.Type, propsJSON,
		graphID, relationID,
	)
	if err != nil {
		return "", fmt.Errorf("mysql: update relation: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "", errors.New("mysql: relation not found")
	}
	return relationID, nil
}

// DeleteRelation 删除关系。
func (s *Store) DeleteRelation(ctx context.Context, graphCode, relationID string) error {
	graphID, err := s.GetGraphIDByCode(ctx, graphCode)
	if err != nil {
		return fmt.Errorf("mysql: get graph id: %w", err)
	}
	if graphID == 0 {
		return nil
	}

	_, err = s.db.ExecContext(ctx,
		`DELETE FROM kg_relation WHERE graph_id = ? AND relation_id = ?`,
		graphID, relationID,
	)
	if err != nil {
		return fmt.Errorf("mysql: delete relation: %w", err)
	}
	return nil
}

// ==================== 内部查询辅助 ====================

func (s *Store) queryNodes(ctx context.Context, graphID int64) ([]*graph.NodeRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT node_id, label, type, chapter_id, summary, properties_json, prerequisites_json, related_json, applies_to_json, targets_json, sort_order
		 FROM kg_node WHERE graph_id = ? ORDER BY sort_order, id`, graphID)
	if err != nil {
		return nil, fmt.Errorf("mysql: query nodes: %w", err)
	}
	defer rows.Close()

	var result []*graph.NodeRow
	for rows.Next() {
		nr := &graph.NodeRow{GraphID: graphID}
		var chapterID sql.NullString
		var summary sql.NullString
		if err := rows.Scan(&nr.NodeID, &nr.Label, &nr.Type, &chapterID, &summary,
			&nr.PropertiesJSON, &nr.PrerequisitesJSON, &nr.RelatedJSON, &nr.AppliesToJSON, &nr.TargetsJSON,
			&nr.SortOrder); err != nil {
			return nil, err
		}
		if chapterID.Valid {
			nr.ChapterID = &chapterID.String
		}
		nr.Summary = summary.String
		result = append(result, nr)
	}
	return result, rows.Err()
}

func (s *Store) queryNodeByID(ctx context.Context, graphID int64, nodeID string) (*graph.NodeRow, error) {
	nr := &graph.NodeRow{GraphID: graphID}
	var chapterID sql.NullString
	var summary sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT node_id, label, type, chapter_id, summary, properties_json, prerequisites_json, related_json, applies_to_json, targets_json, sort_order
		 FROM kg_node WHERE graph_id = ? AND node_id = ?`, graphID, nodeID,
	).Scan(&nr.NodeID, &nr.Label, &nr.Type, &chapterID, &summary,
		&nr.PropertiesJSON, &nr.PrerequisitesJSON, &nr.RelatedJSON, &nr.AppliesToJSON, &nr.TargetsJSON,
		&nr.SortOrder)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if chapterID.Valid {
		nr.ChapterID = &chapterID.String
	}
	nr.Summary = summary.String
	return nr, nil
}

func (s *Store) queryRelations(ctx context.Context, graphID int64) ([]*graph.RelationRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT relation_id, source, target, type, properties_json
		 FROM kg_relation WHERE graph_id = ?`, graphID)
	if err != nil {
		return nil, fmt.Errorf("mysql: query relations: %w", err)
	}
	defer rows.Close()

	var result []*graph.RelationRow
	for rows.Next() {
		rr := &graph.RelationRow{GraphID: graphID}
		if err := rows.Scan(&rr.RelationID, &rr.Source, &rr.Target, &rr.Type, &rr.PropertiesJSON); err != nil {
			return nil, err
		}
		result = append(result, rr)
	}
	return result, rows.Err()
}

func (s *Store) batchInsertNodes(ctx context.Context, tx *sql.Tx, graphID int64, nodes []*graph.Node) (int, error) {
	if len(nodes) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, prerequisites_json, related_json, applies_to_json, targets_json, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("mysql: prepare node insert: %w", err)
	}
	defer stmt.Close()

	count := 0
	for i, node := range nodes {
		if node.ID == "" || node.Label == "" {
			continue
		}
		propsJSON, _ := json.Marshal(node.Properties)
		prereqJSON, _ := json.Marshal(node.Prerequisites)
		relatedJSON, _ := json.Marshal(node.Related)
		appliesJSON, _ := json.Marshal(node.AppliesTo)
		targetsJSON, _ := json.Marshal(node.Targets)

		if _, err := stmt.ExecContext(ctx,
			graphID, node.ID, node.Label, node.Type, node.ChapterID, node.Summary,
			propsJSON, prereqJSON, relatedJSON, appliesJSON, targetsJSON, i,
		); err != nil {
			return 0, fmt.Errorf("mysql: insert node %s: %w", node.ID, err)
		}
		count++
	}
	return count, nil
}

func (s *Store) batchInsertRelations(ctx context.Context, tx *sql.Tx, graphID int64, relations []*graph.Relation) (int, error) {
	if len(relations) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
		 VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("mysql: prepare relation insert: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, rel := range relations {
		if rel.Source == "" || rel.Target == "" || rel.Type == "" {
			continue
		}
		relID := rel.ID
		if relID == "" {
			relID = graph.MakeRelationID(rel.Source, rel.Target, rel.Type)
		}
		propsJSON, _ := json.Marshal(rel.Properties)

		if _, err := stmt.ExecContext(ctx,
			graphID, relID, rel.Source, rel.Target, rel.Type, propsJSON,
		); err != nil {
			return 0, fmt.Errorf("mysql: insert relation %s: %w", relID, err)
		}
		count++
	}
	return count, nil
}

func (s *Store) progressCompletionQuery() (string, []string, string) {
	if s != nil && s.progressCompletionSQL != "" {
		args := s.progressCompletionArgs
		if len(args) == 0 {
			args = []string{"userId"}
		}
		return s.progressCompletionSQL, args, "custom_sql"
	}
	return defaultProgressCompletionSQL, defaultProgressCompletionArgs, "student_assignment"
}

func normalizeProgressArgTokens(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	result := make([]string, 0, len(args))
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			result = append(result, arg)
		}
	}
	return result
}

func buildProgressQueryArgs(tokens []string, graphCode, userID string) ([]any, error) {
	args := make([]any, 0, len(tokens))
	for _, token := range tokens {
		switch strings.ToLower(strings.TrimSpace(token)) {
		case "userid", "user_id", "studentid", "student_id":
			args = append(args, userID)
		case "graphcode", "graph_code":
			args = append(args, graphCode)
		default:
			return nil, fmt.Errorf("mysql: unsupported progress query arg token %q", token)
		}
	}
	return args, nil
}
