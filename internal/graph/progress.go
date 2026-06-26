package graph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	LearningStatusUnstarted = "unstarted"
	LearningStatusLearning  = "learning"
	LearningStatusMastered  = "mastered"
)

const (
	learningStatusUnstartedLabel = "未学习"
	learningStatusLearningLabel  = "学习中"
	learningStatusMasteredLabel  = "已学习"
)

// ComputeProgress 根据已完成练习集合计算图谱中每个节点和全图的学习进度。
func ComputeProgress(graphCode string, g *Graph, userID string, completedExerciseKeys []string, source string) *GraphProgress {
	result := &GraphProgress{
		GraphCode: graphCode,
		UserID:    userID,
		Source:    source,
		Nodes:     map[string]*NodeProgress{},
		UpdatedAt: time.Now(),
	}
	if g == nil {
		status, label := statusForPercent(0)
		result.LearningStatus = status
		result.LearningStatusLabel = label
		return result
	}

	nodes := uniqueNodes(g)
	nodeByID := make(map[string]*Node, len(nodes))
	nodeExercises := make(map[string]map[string]struct{}, len(nodes))
	chapterExercises := map[string]map[string]struct{}{}
	allExercises := map[string]struct{}{}

	for _, node := range nodes {
		nodeByID[node.ID] = node
		if node.Type == "exercise" {
			addSetValue(nodeExercises, node.ID, node.ID)
			allExercises[node.ID] = struct{}{}
			if node.ChapterID != nil {
				addSetValue(chapterExercises, *node.ChapterID, node.ID)
			}
			for _, targetID := range node.Targets {
				addSetValue(nodeExercises, targetID, node.ID)
			}
		}
	}

	for _, rel := range g.Relations {
		if rel == nil || rel.Type != "TESTED_BY" {
			continue
		}
		target, ok := nodeByID[rel.Target]
		if !ok || target.Type != "exercise" {
			continue
		}
		addSetValue(nodeExercises, rel.Source, rel.Target)
		allExercises[rel.Target] = struct{}{}
	}

	for _, node := range nodes {
		if node.ChapterID == nil || node.Type == "chapter" || node.Type == "course" {
			continue
		}
		for exerciseID := range nodeExercises[node.ID] {
			addSetValue(chapterExercises, *node.ChapterID, exerciseID)
		}
	}

	completedExercises := completedExerciseIDSet(nodes, completedExerciseKeys)
	result.CompletedExerciseCount = intersectionCount(allExercises, completedExercises)
	result.TotalExerciseCount = len(allExercises)
	result.ProgressPercent = percent(result.CompletedExerciseCount, result.TotalExerciseCount)
	result.LearningStatus, result.LearningStatusLabel = statusForPercent(result.ProgressPercent)

	for _, node := range nodes {
		var exerciseSet map[string]struct{}
		switch node.Type {
		case "course":
			exerciseSet = allExercises
		case "chapter":
			exerciseSet = chapterExercises[node.ID]
		default:
			exerciseSet = nodeExercises[node.ID]
		}
		result.Nodes[node.ID] = buildNodeProgress(node.ID, node.Type, exerciseSet, completedExercises)
	}

	return result
}

// ApplyProgressToGraph 将用户进度字段附加到图谱节点上，不修改节点结构性字段。
func ApplyProgressToGraph(g *Graph, progress *GraphProgress) {
	if g == nil || progress == nil {
		return
	}
	apply := func(node *Node) {
		if node == nil {
			return
		}
		p, ok := progress.Nodes[node.ID]
		if !ok {
			return
		}
		node.ProgressPercent = intPtr(p.ProgressPercent)
		node.LearningStatus = p.LearningStatus
		node.LearningStatusLabel = p.LearningStatusLabel
		node.CompletedExerciseCount = intPtr(p.CompletedExerciseCount)
		node.TotalExerciseCount = intPtr(p.TotalExerciseCount)
	}
	apply(g.Course)
	for _, node := range g.Nodes {
		apply(node)
	}
}

func uniqueNodes(g *Graph) []*Node {
	if g == nil {
		return nil
	}
	result := make([]*Node, 0, len(g.Nodes)+1)
	seen := map[string]struct{}{}
	add := func(node *Node) {
		if node == nil || node.ID == "" {
			return
		}
		if _, ok := seen[node.ID]; ok {
			return
		}
		seen[node.ID] = struct{}{}
		result = append(result, node)
	}
	add(g.Course)
	for _, node := range g.Nodes {
		add(node)
	}
	return result
}

func completedExerciseIDSet(nodes []*Node, completedExerciseKeys []string) map[string]struct{} {
	keySet := map[string]struct{}{}
	for _, key := range completedExerciseKeys {
		key = strings.ToLower(strings.TrimSpace(key))
		if key != "" {
			keySet[key] = struct{}{}
		}
	}

	result := map[string]struct{}{}
	for _, node := range nodes {
		if node == nil || node.Type != "exercise" {
			continue
		}
		for _, key := range exerciseMatchKeys(node) {
			if _, ok := keySet[key]; ok {
				result[node.ID] = struct{}{}
				break
			}
		}
	}
	return result
}

func exerciseMatchKeys(node *Node) []string {
	keys := []string{strings.ToLower(node.ID)}
	experimentID := mapValueToString(node.Properties["experimentId"])
	if experimentID != "" {
		experimentID = strings.ToLower(experimentID)
		keys = append(keys,
			experimentID,
			"experiment:"+experimentID,
			"experimentid:"+experimentID,
			"ds-ex-"+experimentID,
		)
	}
	return keys
}

func buildNodeProgress(nodeID, nodeType string, exerciseSet map[string]struct{}, completedSet map[string]struct{}) *NodeProgress {
	exerciseIDs := sortedSetValues(exerciseSet)
	completedExerciseIDs := make([]string, 0, len(exerciseIDs))
	for _, exerciseID := range exerciseIDs {
		if _, ok := completedSet[exerciseID]; ok {
			completedExerciseIDs = append(completedExerciseIDs, exerciseID)
		}
	}
	progressPercent := percent(len(completedExerciseIDs), len(exerciseIDs))
	status, label := statusForPercent(progressPercent)
	return &NodeProgress{
		NodeID:                 nodeID,
		NodeType:               nodeType,
		ProgressPercent:        progressPercent,
		LearningStatus:         status,
		LearningStatusLabel:    label,
		CompletedExerciseCount: len(completedExerciseIDs),
		TotalExerciseCount:     len(exerciseIDs),
		ExerciseIDs:            exerciseIDs,
		CompletedExerciseIDs:   completedExerciseIDs,
	}
}

func percent(completed, total int) int {
	if total <= 0 || completed <= 0 {
		return 0
	}
	if completed >= total {
		return 100
	}
	value := completed * 100 / total
	if value == 0 {
		return 1
	}
	return value
}

func statusForPercent(value int) (string, string) {
	if value <= 0 {
		return LearningStatusUnstarted, learningStatusUnstartedLabel
	}
	if value >= 100 {
		return LearningStatusMastered, learningStatusMasteredLabel
	}
	return LearningStatusLearning, learningStatusLearningLabel
}

func addSetValue(target map[string]map[string]struct{}, key, value string) {
	if key == "" || value == "" {
		return
	}
	set, ok := target[key]
	if !ok {
		set = map[string]struct{}{}
		target[key] = set
	}
	set[value] = struct{}{}
}

func sortedSetValues(set map[string]struct{}) []string {
	if len(set) == 0 {
		return []string{}
	}
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func intersectionCount(left, right map[string]struct{}) int {
	count := 0
	for value := range left {
		if _, ok := right[value]; ok {
			count++
		}
	}
	return count
}

func mapValueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func intPtr(value int) *int {
	return &value
}
