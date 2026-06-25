-- 知识图谱服务数据库 Schema
-- 数据库：ptadatabase（与后端/recommendation/spider/LeetCodeClaw 共库）
-- 表前缀：kg_
-- 字符集：utf8mb4

SET NAMES utf8mb4;

-- 图谱元数据表
CREATE TABLE IF NOT EXISTS kg_graph (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    graph_code     VARCHAR(128) NOT NULL COMMENT '唯一标识',
    version        VARCHAR(32)  NOT NULL DEFAULT '1.0.0',
    source_json    JSON COMMENT '来源信息 {system, scenario, origin}',
    metadata_json  JSON COMMENT '元数据 {title, description, domain, ...}',
    course_node_id VARCHAR(128) COMMENT '课程根节点的 node_id',
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_graph_code (graph_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 节点表
CREATE TABLE IF NOT EXISTS kg_node (
    id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    graph_id            BIGINT NOT NULL,
    node_id             VARCHAR(128) NOT NULL COMMENT '业务ID',
    label               VARCHAR(256) NOT NULL COMMENT '显示名称',
    type                VARCHAR(32) NOT NULL COMMENT 'course/chapter/concept/structure/algorithm/operation/exercise',
    chapter_id          VARCHAR(128) COMMENT '所属章节 node_id',
    summary             TEXT COMMENT '简介',
    properties_json     JSON COMMENT '扩展属性 {keywords, studyTip, definition, ...}',
    prerequisites_json  JSON COMMENT '前置节点ID数组',
    related_json        JSON COMMENT '相关节点ID数组',
    applies_to_json     JSON COMMENT '应用于目标ID数组',
    targets_json        JSON COMMENT '练习考察目标ID数组',
    sort_order          INT DEFAULT 0 COMMENT '排序',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_graph_node (graph_id, node_id),
    INDEX idx_graph (graph_id),
    INDEX idx_type (graph_id, type),
    INDEX idx_chapter (graph_id, chapter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 关系表
CREATE TABLE IF NOT EXISTS kg_relation (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    graph_id        BIGINT NOT NULL,
    relation_id     VARCHAR(256) NOT NULL COMMENT '业务ID: source__type__target',
    source          VARCHAR(128) NOT NULL COMMENT '源节点 node_id',
    target          VARCHAR(128) NOT NULL COMMENT '目标节点 node_id',
    type            VARCHAR(32) NOT NULL COMMENT 'CONTAINS/PREREQUISITE/RELATED_TO/APPLIES_TO/TESTED_BY',
    properties_json JSON COMMENT '关系属性',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_graph_relation (graph_id, relation_id),
    INDEX idx_graph (graph_id),
    INDEX idx_source (graph_id, source),
    INDEX idx_target (graph_id, target)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
