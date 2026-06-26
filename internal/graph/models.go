package graph

import (
	"encoding/json"
	"time"
)

// ---- 前端契约数据模型 ----
// 与 frontend-repo/src/features/knowledge-graph/graphDatabaseAdapter.js 对齐。

// Graph 表示一个完整的知识图谱，直接映射前端 normalizeGraph 的输入。
type Graph struct {
	Metadata  MapRaw      `json:"metadata"`
	Course    *Node       `json:"course"`
	Nodes     []*Node     `json:"nodes"`
	Relations []*Relation `json:"relations"`
}

// Node 表示图谱中的一个节点。
type Node struct {
	ID                     string   `json:"id"`
	Label                  string   `json:"label"`
	Type                   string   `json:"type"`
	ChapterID              *string  `json:"chapterId,omitempty"`
	Summary                string   `json:"summary,omitempty"`
	Properties             MapRaw   `json:"properties,omitempty"`
	Prerequisites          []string `json:"prerequisites,omitempty"`
	Related                []string `json:"related,omitempty"`
	AppliesTo              []string `json:"appliesTo,omitempty"`
	Targets                []string `json:"targets,omitempty"`
	ProgressPercent        *int     `json:"progressPercent,omitempty"`
	LearningStatus         string   `json:"learningStatus,omitempty"`
	LearningStatusLabel    string   `json:"learningStatusLabel,omitempty"`
	CompletedExerciseCount *int     `json:"completedExerciseCount,omitempty"`
	TotalExerciseCount     *int     `json:"totalExerciseCount,omitempty"`
}

// Relation 表示图谱中的一条关系。
type Relation struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	Type       string `json:"type"`
	Properties MapRaw `json:"properties,omitempty"`
}

// Payload 表示前端 toGraphDbPayload 的输出格式，用于批量写入。
type Payload struct {
	GraphCode string      `json:"graphCode"`
	Version   string      `json:"version"`
	Source    MapRaw      `json:"source"`
	Metadata  MapRaw      `json:"metadata"`
	Course    *Node       `json:"course"`
	Nodes     []*Node     `json:"nodes"`
	Relations []*Relation `json:"relations"`
}

// GraphMeta 用于图谱列表展示。
type GraphMeta struct {
	GraphCode     string    `json:"graphCode"`
	Version       string    `json:"version"`
	Title         string    `json:"title"`
	NodeCount     int       `json:"nodeCount"`
	RelationCount int       `json:"relationCount"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// NodeContext 表示节点上下文（前端 getNodeContext 对应）。
type NodeContext struct {
	Node          *Node   `json:"node"`
	AncestorChain []*Node `json:"ancestorChain"`
	Prerequisites []*Node `json:"prerequisites"`
	NextNodes     []*Node `json:"nextNodes"`
	RelatedNodes  []*Node `json:"relatedNodes"`
	Exercises     []*Node `json:"exercises"`
}

// GraphStats 表示图谱统计信息。
type GraphStats struct {
	TotalNodes         int            `json:"totalNodes"`
	TotalRelations     int            `json:"totalRelations"`
	NodeTypeCounts     map[string]int `json:"nodeTypeCounts"`
	RelationTypeCounts map[string]int `json:"relationTypeCounts"`
	ChapterCount       int            `json:"chapterCount"`
	StructureCount     int            `json:"structureCount"`
	AlgorithmCount     int            `json:"algorithmCount"`
	ExerciseCount      int            `json:"exerciseCount"`
}

// NodeProgress 表示单个节点的用户学习进度。
type NodeProgress struct {
	NodeID                 string   `json:"nodeId"`
	NodeType               string   `json:"nodeType"`
	ProgressPercent        int      `json:"progressPercent"`
	LearningStatus         string   `json:"learningStatus"`
	LearningStatusLabel    string   `json:"learningStatusLabel"`
	CompletedExerciseCount int      `json:"completedExerciseCount"`
	TotalExerciseCount     int      `json:"totalExerciseCount"`
	ExerciseIDs            []string `json:"exerciseIds,omitempty"`
	CompletedExerciseIDs   []string `json:"completedExerciseIds,omitempty"`
}

// GraphProgress 表示某个用户在图谱上的聚合学习进度。
type GraphProgress struct {
	GraphCode              string                   `json:"graphCode"`
	UserID                 string                   `json:"userId"`
	ProgressPercent        int                      `json:"progressPercent"`
	LearningStatus         string                   `json:"learningStatus"`
	LearningStatusLabel    string                   `json:"learningStatusLabel"`
	CompletedExerciseCount int                      `json:"completedExerciseCount"`
	TotalExerciseCount     int                      `json:"totalExerciseCount"`
	Source                 string                   `json:"source"`
	Nodes                  map[string]*NodeProgress `json:"nodes"`
	UpdatedAt              time.Time                `json:"updatedAt"`
}

// ValidationResult 表示图谱校验结果。
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// ---- 内部数据库行模型 ----

// GraphRow 映射 kg_graph 表。
type GraphRow struct {
	ID           int64     `json:"-"`
	GraphCode    string    `json:"graphCode"`
	Version      string    `json:"version"`
	SourceJSON   []byte    `json:"-"`
	MetadataJSON []byte    `json:"-"`
	CourseNodeID *string   `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// NodeRow 映射 kg_node 表。
type NodeRow struct {
	ID                int64   `json:"-"`
	GraphID           int64   `json:"-"`
	NodeID            string  `json:"nodeId"`
	Label             string  `json:"label"`
	Type              string  `json:"type"`
	ChapterID         *string `json:"chapterId,omitempty"`
	Summary           string  `json:"summary,omitempty"`
	PropertiesJSON    []byte  `json:"-"`
	PrerequisitesJSON []byte  `json:"-"`
	RelatedJSON       []byte  `json:"-"`
	AppliesToJSON     []byte  `json:"-"`
	TargetsJSON       []byte  `json:"-"`
	SortOrder         int     `json:"-"`
}

// RelationRow 映射 kg_relation 表。
type RelationRow struct {
	ID             int64  `json:"-"`
	GraphID        int64  `json:"-"`
	RelationID     string `json:"relationId"`
	Source         string `json:"source"`
	Target         string `json:"target"`
	Type           string `json:"type"`
	PropertiesJSON []byte `json:"-"`
}

// ---- 辅助类型 ----

// MapRaw 是一个保留任意键值的 JSON 兼容类型。
type MapRaw map[string]any

// MarshalJSON 让 MapRaw 在非空时正常序列化，空时输出 {}。
func (m MapRaw) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]any(m))
}

// ---- 行模型 → 前端模型 转换 ----

// RowToNode 将数据库行转换为前端 Node。
func RowToNode(row *NodeRow) *Node {
	node := &Node{
		ID:         row.NodeID,
		Label:      row.Label,
		Type:       row.Type,
		ChapterID:  row.ChapterID,
		Summary:    row.Summary,
		Properties: ParseMapRaw(row.PropertiesJSON),
	}
	node.Prerequisites = parseStringSlice(row.PrerequisitesJSON)
	node.Related = parseStringSlice(row.RelatedJSON)
	node.AppliesTo = parseStringSlice(row.AppliesToJSON)
	node.Targets = parseStringSlice(row.TargetsJSON)
	return node
}

// RowToRelation 将数据库行转换为前端 Relation。
func RowToRelation(row *RelationRow) *Relation {
	return &Relation{
		ID:         row.RelationID,
		Source:     row.Source,
		Target:     row.Target,
		Type:       row.Type,
		Properties: ParseMapRaw(row.PropertiesJSON),
	}
}

// ---- JSON 解析辅助 ----

// ParseMapRaw 将 JSON 字节解析为 MapRaw。
func ParseMapRaw(data []byte) MapRaw {
	if len(data) == 0 {
		return MapRaw{}
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return MapRaw{}
	}
	return MapRaw(m)
}

func parseStringSlice(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return []string{}
	}
	return arr
}

// MakeRelationID 生成关系业务ID: source__type__target
func MakeRelationID(source, target, relType string) string {
	return source + "__" + relType + "__" + target
}
