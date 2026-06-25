package graph

// 合法节点类型和关系类型，与前端 dataStructureGraph.js 对齐。

var validNodeTypes = map[string]bool{
	"course":    true,
	"chapter":   true,
	"concept":   true,
	"structure": true,
	"algorithm": true,
	"operation": true,
	"exercise":  true,
}

var validRelationTypes = map[string]bool{
	"CONTAINS":     true,
	"PREREQUISITE": true,
	"RELATED_TO":   true,
	"APPLIES_TO":   true,
	"TESTED_BY":    true,
}

// IsValidNodeType 检查节点类型是否合法。
func IsValidNodeType(t string) bool {
	return validNodeTypes[t]
}

// IsValidRelationType 检查关系类型是否合法。
func IsValidRelationType(t string) bool {
	return validRelationTypes[t]
}

// ValidateGraph 对完整图谱执行校验，与前端 validateGraph 逻辑一致。
func ValidateGraph(g *Graph) *ValidationResult {
	errors := []string{}
	if g == nil {
		return &ValidationResult{Valid: false, Errors: []string{"图谱为空"}}
	}

	nodeIDs := make(map[string]bool)
	nodeIDList := make([]string, 0, len(g.Nodes))

	// 把 course 也加入节点列表
	allNodes := make([]*Node, 0, len(g.Nodes)+1)
	if g.Course != nil && g.Course.ID != "" {
		allNodes = append(allNodes, g.Course)
	}
	allNodes = append(allNodes, g.Nodes...)

	// 检查重复节点 ID
	for _, node := range allNodes {
		if node == nil || node.ID == "" {
			continue
		}
		if nodeIDs[node.ID] {
			errors = append(errors, "重复的节点 id: "+node.ID)
			continue
		}
		nodeIDs[node.ID] = true
		nodeIDList = append(nodeIDList, node.ID)
	}

	// 检查重复关系 ID 和关系端点存在性
	relationIDs := make(map[string]bool)
	for _, rel := range g.Relations {
		if rel == nil {
			continue
		}
		if rel.ID == "" {
			rel.ID = MakeRelationID(rel.Source, rel.Target, rel.Type)
		}
		if relationIDs[rel.ID] {
			errors = append(errors, "重复的关系 id: "+rel.ID)
			continue
		}
		relationIDs[rel.ID] = true

		if !nodeIDs[rel.Source] {
			errors = append(errors, "关系 "+rel.ID+" 的 source 不存在: "+rel.Source)
		}
		if !nodeIDs[rel.Target] {
			errors = append(errors, "关系 "+rel.ID+" 的 target 不存在: "+rel.Target)
		}
		if !IsValidRelationType(rel.Type) {
			errors = append(errors, "非法关系类型: "+rel.Type)
		}
	}

	// 检查课程根节点
	courseExists := false
	if g.Course != nil && g.Course.ID != "" && nodeIDs[g.Course.ID] {
		courseExists = true
	}
	if !courseExists {
		// 检查 nodes 中是否有 course 类型
		for _, node := range allNodes {
			if node != nil && node.Type == "course" && node.ID != "" {
				courseExists = true
				break
			}
		}
	}
	if !courseExists {
		errors = append(errors, "课程根节点不存在")
	}

	// 检查至少一个章节节点
	hasChapter := false
	for _, node := range allNodes {
		if node != nil && node.Type == "chapter" {
			hasChapter = true
			break
		}
	}
	if !hasChapter {
		errors = append(errors, "至少需要一个章节节点")
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ComputeStats 计算图谱统计信息，与前端 getGraphStats 逻辑一致。
func ComputeStats(g *Graph) *GraphStats {
	stats := &GraphStats{
		NodeTypeCounts:     make(map[string]int),
		RelationTypeCounts: make(map[string]int),
	}
	if g == nil {
		return stats
	}

	allNodes := make([]*Node, 0, len(g.Nodes)+1)
	if g.Course != nil {
		allNodes = append(allNodes, g.Course)
	}
	allNodes = append(allNodes, g.Nodes...)

	for _, node := range allNodes {
		if node == nil {
			continue
		}
		stats.TotalNodes++
		stats.NodeTypeCounts[node.Type]++
	}

	for _, rel := range g.Relations {
		if rel == nil {
			continue
		}
		stats.TotalRelations++
		stats.RelationTypeCounts[rel.Type]++
	}

	stats.ChapterCount = stats.NodeTypeCounts["chapter"]
	stats.StructureCount = stats.NodeTypeCounts["structure"]
	stats.AlgorithmCount = stats.NodeTypeCounts["algorithm"]
	stats.ExerciseCount = stats.NodeTypeCounts["exercise"]

	return stats
}
