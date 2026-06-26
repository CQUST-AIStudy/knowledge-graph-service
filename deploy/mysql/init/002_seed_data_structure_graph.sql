-- 数据结构课程知识图谱种子数据
-- graph_code: data-structure-knowledge-graph（与前端 GRAPH_CODE 常量一致）
-- 基于 SkillTreeConfig 的 6 个维度和 19 个实验

SET NAMES utf8mb4;

-- 幂等：已存在则跳过
INSERT IGNORE INTO kg_graph (graph_code, version, source_json, metadata_json, course_node_id)
VALUES (
    'data-structure-knowledge-graph',
    '1.3.0',
    JSON_OBJECT('system', 'knowledge-graph-service', 'scenario', 'course-chapter-knowledge-graph', 'origin', 'seed-sql'),
    JSON_OBJECT(
        'title', '数据结构课程知识图谱',
        'description', '覆盖线性表、栈与队列、树、图、哈希等核心数据结构的知识体系',
        'domain', 'computer-science',
        'architecture', 'course-chapter',
        'audience', 'student-and-teacher'
    ),
    'ds-course'
);

SET @gid = (SELECT id FROM kg_graph WHERE graph_code = 'data-structure-knowledge-graph');

-- ==================== 课程节点 ====================
INSERT IGNORE INTO kg_node (graph_id, node_id, label, type, summary, properties_json, sort_order)
VALUES (@gid, 'ds-course', '数据结构', 'course', '计算机科学核心基础课程，研究数据的逻辑结构、存储结构及其运算', JSON_OBJECT('keywords', JSON_ARRAY('数据结构','算法','复杂度'), 'studyTip', '从线性到非线性，从静态到动态，循序渐进'), 0);

-- ==================== 章节节点（6个） ====================
INSERT IGNORE INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, sort_order)
VALUES
(@gid, 'ds-ch-linear', '线性表', 'chapter', 'ds-course', '顺序表、单链表、双向链表、循环链表等线性数据结构', JSON_OBJECT('keywords', JSON_ARRAY('线性表','顺序表','链表'), 'difficulty', '基础'), 1),
(@gid, 'ds-ch-stack-queue', '栈与队列', 'chapter', 'ds-course', '栈的实现与应用、队列的实现', JSON_OBJECT('keywords', JSON_ARRAY('栈','队列','LIFO','FIFO'), 'difficulty', '基础'), 2),
(@gid, 'ds-ch-tree', '树', 'chapter', 'ds-course', '二叉搜索树、二叉树遍历、Huffman树', JSON_OBJECT('keywords', JSON_ARRAY('树','二叉树','BST','Huffman'), 'difficulty', '进阶'), 3),
(@gid, 'ds-ch-graph', '图', 'chapter', 'ds-course', 'DFS/BFS、Dijkstra/Prim最短路径与最小生成树', JSON_OBJECT('keywords', JSON_ARRAY('图','DFS','BFS','Dijkstra','Prim'), 'difficulty', '进阶'), 4),
(@gid, 'ds-ch-hash', '哈希', 'chapter', 'ds-course', '哈希表的实现与冲突处理', JSON_OBJECT('keywords', JSON_ARRAY('哈希','哈希表','冲突'), 'difficulty', '进阶'), 5),
(@gid, 'ds-ch-comp', '综合', 'chapter', 'ds-course', '综合练习与期中复习', JSON_OBJECT('keywords', JSON_ARRAY('综合','复习'), 'difficulty', '综合'), 6);

-- ==================== 结构节点 ====================
-- 线性表结构
INSERT IGNORE INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, sort_order)
VALUES
(@gid, 'ds-struct-seqlist', '顺序表', 'structure', 'ds-ch-linear', '基于数组实现的线性表，支持随机访问', JSON_OBJECT('keywords', JSON_ARRAY('数组','随机访问','O(1)访问'), 'estimatedMinutes', 30, 'difficulty', '基础'), 10),
(@gid, 'ds-struct-linkedlist', '单链表', 'structure', 'ds-ch-linear', '基于指针实现的线性表，动态分配内存', JSON_OBJECT('keywords', JSON_ARRAY('指针','节点','动态分配'), 'estimatedMinutes', 45, 'difficulty', '基础'), 11),
(@gid, 'ds-struct-dlist', '双向链表', 'structure', 'ds-ch-linear', '每个节点含前驱和后继指针，支持双向遍历', JSON_OBJECT('keywords', JSON_ARRAY('双向','前驱','后继'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 12),
(@gid, 'ds-struct-clist', '循环链表', 'structure', 'ds-ch-linear', '尾节点指向头节点，形成环结构', JSON_OBJECT('keywords', JSON_ARRAY('循环','环'), 'estimatedMinutes', 20, 'difficulty', '进阶'), 13),

-- 栈与队列结构
(@gid, 'ds-struct-seqstack', '顺序栈', 'structure', 'ds-ch-stack-queue', '基于数组实现的栈，后进先出 LIFO', JSON_OBJECT('keywords', JSON_ARRAY('栈','LIFO','数组'), 'estimatedMinutes', 30, 'difficulty', '基础'), 20),
(@gid, 'ds-struct-linkstack', '链栈', 'structure', 'ds-ch-stack-queue', '基于链表实现的栈', JSON_OBJECT('keywords', JSON_ARRAY('栈','LIFO','链表'), 'estimatedMinutes', 20, 'difficulty', '基础'), 21),
(@gid, 'ds-struct-seqqueue', '顺序队列', 'structure', 'ds-ch-stack-queue', '基于数组实现的队列，先进先出 FIFO', JSON_OBJECT('keywords', JSON_ARRAY('队列','FIFO','数组'), 'estimatedMinutes', 30, 'difficulty', '基础'), 22),
(@gid, 'ds-struct-circqueue', '循环队列', 'structure', 'ds-ch-stack-queue', '利用取模运算实现环形队列，避免假溢出', JSON_OBJECT('keywords', JSON_ARRAY('循环队列','取模','环形'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 23),

-- 树结构
(@gid, 'ds-struct-btree', '二叉树', 'structure', 'ds-ch-tree', '每个节点最多两个子树的树结构', JSON_OBJECT('keywords', JSON_ARRAY('二叉树','左子树','右子树'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 30),
(@gid, 'ds-struct-bst', '二叉搜索树', 'structure', 'ds-ch-tree', '左子节点小于根，右子节点大于根的二叉树', JSON_OBJECT('keywords', JSON_ARRAY('BST','搜索','有序'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 31),
(@gid, 'ds-struct-huffman', 'Huffman树', 'structure', 'ds-ch-tree', '带权路径长度最小的二叉树，用于数据压缩', JSON_OBJECT('keywords', JSON_ARRAY('Huffman','编码','压缩','带权路径'), 'estimatedMinutes', 60, 'difficulty', '进阶'), 32),

-- 哈希结构
(@gid, 'ds-struct-hashtable', '哈希表', 'structure', 'ds-ch-hash', '通过哈希函数将键映射到数组位置的数据结构', JSON_OBJECT('keywords', JSON_ARRAY('哈希表','哈希函数','桶'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 40),

-- 图结构
(@gid, 'ds-struct-adjmatrix', '邻接矩阵', 'structure', 'ds-ch-graph', '用二维数组表示图中顶点间关系的存储方法', JSON_OBJECT('keywords', JSON_ARRAY('邻接矩阵','二维数组','稠密图'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 50),
(@gid, 'ds-struct-adjlist', '邻接表', 'structure', 'ds-ch-graph', '用链表数组表示图中顶点间关系的存储方法', JSON_OBJECT('keywords', JSON_ARRAY('邻接表','链表','稀疏图'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 51);

-- ==================== 算法节点 ====================
-- 线性表算法
INSERT IGNORE INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, sort_order)
VALUES
(@gid, 'ds-algo-insert', '插入', 'algorithm', 'ds-ch-linear', '在线性表的指定位置插入新元素', JSON_OBJECT('keywords', JSON_ARRAY('插入','O(n)移动'), 'estimatedMinutes', 20, 'difficulty', '基础'), 14),
(@gid, 'ds-algo-delete', '删除', 'algorithm', 'ds-ch-linear', '删除线性表中指定位置的元素', JSON_OBJECT('keywords', JSON_ARRAY('删除','O(n)移动'), 'estimatedMinutes', 20, 'difficulty', '基础'), 15),
(@gid, 'ds-algo-search', '查找', 'algorithm', 'ds-ch-linear', '在线性表中查找满足条件的元素', JSON_OBJECT('keywords', JSON_ARRAY('查找','顺序查找','二分查找'), 'estimatedMinutes', 20, 'difficulty', '基础'), 16),

-- 栈与队列算法
(@gid, 'ds-algo-pushpop', '入栈出栈', 'algorithm', 'ds-ch-stack-queue', '压入元素到栈顶、弹出栈顶元素', JSON_OBJECT('keywords', JSON_ARRAY('push','pop','LIFO'), 'estimatedMinutes', 15, 'difficulty', '基础'), 24),
(@gid, 'ds-algo-enqdeq', '入队出队', 'algorithm', 'ds-ch-stack-queue', '在队尾入队、在队头出队', JSON_OBJECT('keywords', JSON_ARRAY('enqueue','dequeue','FIFO'), 'estimatedMinutes', 15, 'difficulty', '基础'), 25),

-- 树算法
(@gid, 'ds-algo-preorder', '前序遍历', 'algorithm', 'ds-ch-tree', '根→左→右的遍历顺序', JSON_OBJECT('keywords', JSON_ARRAY('前序','根左右','递归'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 33),
(@gid, 'ds-algo-inorder', '中序遍历', 'algorithm', 'ds-ch-tree', '左→根→右的遍历顺序，BST中序遍历得到有序序列', JSON_OBJECT('keywords', JSON_ARRAY('中序','左根右','有序'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 34),
(@gid, 'ds-algo-postorder', '后序遍历', 'algorithm', 'ds-ch-tree', '左→右→根的遍历顺序', JSON_OBJECT('keywords', JSON_ARRAY('后序','左右根'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 35),
(@gid, 'ds-algo-levelorder', '层序遍历', 'algorithm', 'ds-ch-tree', '按层从上到下、从左到右遍历，使用队列辅助', JSON_OBJECT('keywords', JSON_ARRAY('层序','队列','BFS'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 36),
(@gid, 'ds-algo-huffman-encode', 'Huffman编码', 'algorithm', 'ds-ch-tree', '根据Huffman树生成前缀编码', JSON_OBJECT('keywords', JSON_ARRAY('Huffman','编码','压缩'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 37),

-- 哈希算法
(@gid, 'ds-algo-hashfunc', '哈希函数', 'algorithm', 'ds-ch-hash', '将键映射到数组索引的函数，如除留余数法', JSON_OBJECT('keywords', JSON_ARRAY('哈希函数','除留余数','映射'), 'estimatedMinutes', 20, 'difficulty', '进阶'), 41),
(@gid, 'ds-algo-collision', '冲突处理', 'algorithm', 'ds-ch-hash', '处理不同键映射到相同位置的方法：链地址法、开放地址法', JSON_OBJECT('keywords', JSON_ARRAY('冲突','链地址法','开放地址法','线性探测'), 'estimatedMinutes', 30, 'difficulty', '进阶'), 42),

-- 图算法
(@gid, 'ds-algo-dfs', '深度优先搜索', 'algorithm', 'ds-ch-graph', '沿一条路径深入到底再回溯的图遍历算法', JSON_OBJECT('keywords', JSON_ARRAY('DFS','深度优先','递归','栈'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 52),
(@gid, 'ds-algo-bfs', '广度优先搜索', 'algorithm', 'ds-ch-graph', '按层扩展的图遍历算法，使用队列辅助', JSON_OBJECT('keywords', JSON_ARRAY('BFS','广度优先','队列'), 'estimatedMinutes', 45, 'difficulty', '进阶'), 53),
(@gid, 'ds-algo-dijkstra', 'Dijkstra算法', 'algorithm', 'ds-ch-graph', '求单源最短路径的贪心算法', JSON_OBJECT('keywords', JSON_ARRAY('Dijkstra','最短路径','贪心'), 'estimatedMinutes', 60, 'difficulty', '进阶'), 54),
(@gid, 'ds-algo-prim', 'Prim算法', 'algorithm', 'ds-ch-graph', '求最小生成树的贪心算法', JSON_OBJECT('keywords', JSON_ARRAY('Prim','最小生成树','贪心'), 'estimatedMinutes', 60, 'difficulty', '进阶'), 55);

-- ==================== 练习节点（19个实验） ====================
INSERT IGNORE INTO kg_node (graph_id, node_id, label, type, chapter_id, summary, properties_json, targets_json, sort_order)
VALUES
(@gid, 'ds-ex-1', '第1次作业', 'exercise', 'ds-ch-linear', '线性表基础作业', JSON_OBJECT('experimentId', 1), JSON_ARRAY('ds-struct-seqlist'), 100),
(@gid, 'ds-ex-2', '第1次实验', 'exercise', 'ds-ch-linear', '顺序表实验', JSON_OBJECT('experimentId', 2), JSON_ARRAY('ds-struct-seqlist','ds-algo-insert','ds-algo-delete'), 101),
(@gid, 'ds-ex-3', '第2次作业(单链表)', 'exercise', 'ds-ch-linear', '单链表作业', JSON_OBJECT('experimentId', 3), JSON_ARRAY('ds-struct-linkedlist'), 102),
(@gid, 'ds-ex-4', '第2次实验(单链表)', 'exercise', 'ds-ch-linear', '单链表实验', JSON_OBJECT('experimentId', 4), JSON_ARRAY('ds-struct-linkedlist','ds-algo-insert','ds-algo-delete'), 103),
(@gid, 'ds-ex-5', '第3次作业(单链表)', 'exercise', 'ds-ch-linear', '单链表进阶作业', JSON_OBJECT('experimentId', 5), JSON_ARRAY('ds-struct-linkedlist','ds-algo-search'), 104),
(@gid, 'ds-ex-6', '第3次实验(链表应用)', 'exercise', 'ds-ch-linear', '链表应用实验', JSON_OBJECT('experimentId', 6), JSON_ARRAY('ds-struct-dlist','ds-struct-clist'), 105),
(@gid, 'ds-ex-7', '第4次作业(双向循环链表)', 'exercise', 'ds-ch-linear', '双向循环链表作业', JSON_OBJECT('experimentId', 7), JSON_ARRAY('ds-struct-dlist','ds-struct-clist'), 106),
(@gid, 'ds-ex-8', '第4次实验(栈)', 'exercise', 'ds-ch-stack-queue', '顺序栈实验', JSON_OBJECT('experimentId', 8), JSON_ARRAY('ds-struct-seqstack','ds-algo-pushpop'), 107),
(@gid, 'ds-ex-9', '第5次实验(队列)', 'exercise', 'ds-ch-stack-queue', '循环队列实验', JSON_OBJECT('experimentId', 9), JSON_ARRAY('ds-struct-circqueue','ds-algo-enqdeq'), 108),
(@gid, 'ds-ex-10', '第6次作业(BST)', 'exercise', 'ds-ch-tree', '二叉搜索树作业', JSON_OBJECT('experimentId', 10), JSON_ARRAY('ds-struct-bst','ds-algo-search'), 109),
(@gid, 'ds-ex-11', '第6次实验(二叉树遍历)', 'exercise', 'ds-ch-tree', '二叉树遍历实验', JSON_OBJECT('experimentId', 11), JSON_ARRAY('ds-struct-btree','ds-algo-preorder','ds-algo-inorder','ds-algo-postorder','ds-algo-levelorder'), 110),
(@gid, 'ds-ex-12', '第7次实验(Huffman)', 'exercise', 'ds-ch-tree', 'Huffman树与编码实验', JSON_OBJECT('experimentId', 12), JSON_ARRAY('ds-struct-huffman','ds-algo-huffman-encode'), 111),
(@gid, 'ds-ex-13', '第8次实验(HashTable)', 'exercise', 'ds-ch-hash', '哈希表实验', JSON_OBJECT('experimentId', 13), JSON_ARRAY('ds-struct-hashtable','ds-algo-hashfunc','ds-algo-collision'), 112),
(@gid, 'ds-ex-14', '第9次实验(DFS/BFS)', 'exercise', 'ds-ch-graph', '图的DFS与BFS实验', JSON_OBJECT('experimentId', 14), JSON_ARRAY('ds-struct-adjmatrix','ds-struct-adjlist','ds-algo-dfs','ds-algo-bfs'), 113),
(@gid, 'ds-ex-15', '第10次实验(栈应用)', 'exercise', 'ds-ch-stack-queue', '栈应用实验', JSON_OBJECT('experimentId', 15), JSON_ARRAY('ds-struct-seqstack','ds-algo-pushpop'), 114),
(@gid, 'ds-ex-16', '第11次实验(Dijkstra/Prim)', 'exercise', 'ds-ch-graph', '最短路径与最小生成树实验', JSON_OBJECT('experimentId', 16), JSON_ARRAY('ds-algo-dijkstra','ds-algo-prim'), 115),
(@gid, 'ds-ex-17', '第12次实验', 'exercise', 'ds-ch-comp', '综合实验', JSON_OBJECT('experimentId', 17), JSON_ARRAY('ds-struct-seqlist','ds-struct-linkedlist'), 116),
(@gid, 'ds-ex-18', '期中复习', 'exercise', 'ds-ch-comp', '期中综合复习', JSON_OBJECT('experimentId', 18), JSON_ARRAY(), 117),
(@gid, 'ds-ex-19', '例题', 'exercise', 'ds-ch-comp', '综合例题', JSON_OBJECT('experimentId', 19), JSON_ARRAY(), 118);

-- ==================== CONTAINS 关系 ====================
-- 课程→章节
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
SELECT @gid, CONCAT('ds-course__CONTAINS__', node_id), 'ds-course', node_id, 'CONTAINS', JSON_OBJECT('scope','course-chapter')
FROM kg_node WHERE graph_id = @gid AND type = 'chapter';

-- 章节→结构/算法/练习
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
SELECT @gid, CONCAT(chapter_id, '__CONTAINS__', node_id), chapter_id, node_id, 'CONTAINS', JSON_OBJECT('scope','chapter-node')
FROM kg_node WHERE graph_id = @gid AND chapter_id IS NOT NULL AND type IN ('structure','algorithm','exercise');

-- ==================== PREREQUISITE 关系 ====================
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
VALUES
-- 章节间前置
(@gid, 'ds-ch-linear__PREREQUISITE__ds-ch-stack-queue', 'ds-ch-linear', 'ds-ch-stack-queue', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-ch-stack-queue__PREREQUISITE__ds-ch-tree', 'ds-ch-stack-queue', 'ds-ch-tree', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-ch-tree__PREREQUISITE__ds-ch-graph', 'ds-ch-tree', 'ds-ch-graph', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-ch-tree__PREREQUISITE__ds-ch-hash', 'ds-ch-tree', 'ds-ch-hash', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
-- 线性表内前置
(@gid, 'ds-struct-seqlist__PREREQUISITE__ds-struct-linkedlist', 'ds-struct-seqlist', 'ds-struct-linkedlist', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-struct-linkedlist__PREREQUISITE__ds-struct-dlist', 'ds-struct-linkedlist', 'ds-struct-dlist', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-struct-linkedlist__PREREQUISITE__ds-struct-clist', 'ds-struct-linkedlist', 'ds-struct-clist', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
-- 栈与队列内前置
(@gid, 'ds-struct-seqstack__PREREQUISITE__ds-struct-linkstack', 'ds-struct-seqstack', 'ds-struct-linkstack', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-struct-seqqueue__PREREQUISITE__ds-struct-circqueue', 'ds-struct-seqqueue', 'ds-struct-circqueue', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
-- 树内前置
(@gid, 'ds-struct-btree__PREREQUISITE__ds-struct-bst', 'ds-struct-btree', 'ds-struct-bst', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order')),
(@gid, 'ds-struct-bst__PREREQUISITE__ds-struct-huffman', 'ds-struct-bst', 'ds-struct-huffman', 'PREREQUISITE', JSON_OBJECT('kind','knowledge-order'));

-- ==================== APPLIES_TO 关系（算法→结构） ====================
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
VALUES
(@gid, 'ds-algo-insert__APPLIES_TO__ds-struct-seqlist', 'ds-algo-insert', 'ds-struct-seqlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-insert__APPLIES_TO__ds-struct-linkedlist', 'ds-algo-insert', 'ds-struct-linkedlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-delete__APPLIES_TO__ds-struct-seqlist', 'ds-algo-delete', 'ds-struct-seqlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-delete__APPLIES_TO__ds-struct-linkedlist', 'ds-algo-delete', 'ds-struct-linkedlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-search__APPLIES_TO__ds-struct-seqlist', 'ds-algo-search', 'ds-struct-seqlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-search__APPLIES_TO__ds-struct-linkedlist', 'ds-algo-search', 'ds-struct-linkedlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-search__APPLIES_TO__ds-struct-bst', 'ds-algo-search', 'ds-struct-bst', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-pushpop__APPLIES_TO__ds-struct-seqstack', 'ds-algo-pushpop', 'ds-struct-seqstack', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-pushpop__APPLIES_TO__ds-struct-linkstack', 'ds-algo-pushpop', 'ds-struct-linkstack', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-enqdeq__APPLIES_TO__ds-struct-seqqueue', 'ds-algo-enqdeq', 'ds-struct-seqqueue', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-enqdeq__APPLIES_TO__ds-struct-circqueue', 'ds-algo-enqdeq', 'ds-struct-circqueue', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-preorder__APPLIES_TO__ds-struct-btree', 'ds-algo-preorder', 'ds-struct-btree', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-inorder__APPLIES_TO__ds-struct-btree', 'ds-algo-inorder', 'ds-struct-btree', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-postorder__APPLIES_TO__ds-struct-btree', 'ds-algo-postorder', 'ds-struct-btree', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-levelorder__APPLIES_TO__ds-struct-btree', 'ds-algo-levelorder', 'ds-struct-btree', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-huffman-encode__APPLIES_TO__ds-struct-huffman', 'ds-algo-huffman-encode', 'ds-struct-huffman', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-hashfunc__APPLIES_TO__ds-struct-hashtable', 'ds-algo-hashfunc', 'ds-struct-hashtable', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-collision__APPLIES_TO__ds-struct-hashtable', 'ds-algo-collision', 'ds-struct-hashtable', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-dfs__APPLIES_TO__ds-struct-adjmatrix', 'ds-algo-dfs', 'ds-struct-adjmatrix', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-dfs__APPLIES_TO__ds-struct-adjlist', 'ds-algo-dfs', 'ds-struct-adjlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-bfs__APPLIES_TO__ds-struct-adjmatrix', 'ds-algo-bfs', 'ds-struct-adjmatrix', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-bfs__APPLIES_TO__ds-struct-adjlist', 'ds-algo-bfs', 'ds-struct-adjlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-dijkstra__APPLIES_TO__ds-struct-adjmatrix', 'ds-algo-dijkstra', 'ds-struct-adjmatrix', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-dijkstra__APPLIES_TO__ds-struct-adjlist', 'ds-algo-dijkstra', 'ds-struct-adjlist', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-prim__APPLIES_TO__ds-struct-adjmatrix', 'ds-algo-prim', 'ds-struct-adjmatrix', 'APPLIES_TO', JSON_OBJECT('kind','application')),
(@gid, 'ds-algo-prim__APPLIES_TO__ds-struct-adjlist', 'ds-algo-prim', 'ds-struct-adjlist', 'APPLIES_TO', JSON_OBJECT('kind','application'));

-- ==================== TESTED_BY 关系（结构/算法→练习） ====================
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
VALUES
(@gid, 'ds-struct-seqlist__TESTED_BY__ds-ex-1', 'ds-struct-seqlist', 'ds-ex-1', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-seqlist__TESTED_BY__ds-ex-2', 'ds-struct-seqlist', 'ds-ex-2', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-insert__TESTED_BY__ds-ex-2', 'ds-algo-insert', 'ds-ex-2', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-delete__TESTED_BY__ds-ex-2', 'ds-algo-delete', 'ds-ex-2', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-linkedlist__TESTED_BY__ds-ex-3', 'ds-struct-linkedlist', 'ds-ex-3', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-linkedlist__TESTED_BY__ds-ex-4', 'ds-struct-linkedlist', 'ds-ex-4', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-insert__TESTED_BY__ds-ex-4', 'ds-algo-insert', 'ds-ex-4', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-delete__TESTED_BY__ds-ex-4', 'ds-algo-delete', 'ds-ex-4', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-linkedlist__TESTED_BY__ds-ex-5', 'ds-struct-linkedlist', 'ds-ex-5', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-search__TESTED_BY__ds-ex-5', 'ds-algo-search', 'ds-ex-5', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-dlist__TESTED_BY__ds-ex-6', 'ds-struct-dlist', 'ds-ex-6', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-clist__TESTED_BY__ds-ex-6', 'ds-struct-clist', 'ds-ex-6', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-dlist__TESTED_BY__ds-ex-7', 'ds-struct-dlist', 'ds-ex-7', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-clist__TESTED_BY__ds-ex-7', 'ds-struct-clist', 'ds-ex-7', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-seqstack__TESTED_BY__ds-ex-8', 'ds-struct-seqstack', 'ds-ex-8', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-pushpop__TESTED_BY__ds-ex-8', 'ds-algo-pushpop', 'ds-ex-8', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-circqueue__TESTED_BY__ds-ex-9', 'ds-struct-circqueue', 'ds-ex-9', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-enqdeq__TESTED_BY__ds-ex-9', 'ds-algo-enqdeq', 'ds-ex-9', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-bst__TESTED_BY__ds-ex-10', 'ds-struct-bst', 'ds-ex-10', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-search__TESTED_BY__ds-ex-10', 'ds-algo-search', 'ds-ex-10', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-btree__TESTED_BY__ds-ex-11', 'ds-struct-btree', 'ds-ex-11', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-preorder__TESTED_BY__ds-ex-11', 'ds-algo-preorder', 'ds-ex-11', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-inorder__TESTED_BY__ds-ex-11', 'ds-algo-inorder', 'ds-ex-11', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-postorder__TESTED_BY__ds-ex-11', 'ds-algo-postorder', 'ds-ex-11', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-levelorder__TESTED_BY__ds-ex-11', 'ds-algo-levelorder', 'ds-ex-11', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-huffman__TESTED_BY__ds-ex-12', 'ds-struct-huffman', 'ds-ex-12', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-huffman-encode__TESTED_BY__ds-ex-12', 'ds-algo-huffman-encode', 'ds-ex-12', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-hashtable__TESTED_BY__ds-ex-13', 'ds-struct-hashtable', 'ds-ex-13', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-hashfunc__TESTED_BY__ds-ex-13', 'ds-algo-hashfunc', 'ds-ex-13', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-collision__TESTED_BY__ds-ex-13', 'ds-algo-collision', 'ds-ex-13', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-adjmatrix__TESTED_BY__ds-ex-14', 'ds-struct-adjmatrix', 'ds-ex-14', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-adjlist__TESTED_BY__ds-ex-14', 'ds-struct-adjlist', 'ds-ex-14', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-dfs__TESTED_BY__ds-ex-14', 'ds-algo-dfs', 'ds-ex-14', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-bfs__TESTED_BY__ds-ex-14', 'ds-algo-bfs', 'ds-ex-14', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-seqstack__TESTED_BY__ds-ex-15', 'ds-struct-seqstack', 'ds-ex-15', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-pushpop__TESTED_BY__ds-ex-15', 'ds-algo-pushpop', 'ds-ex-15', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-dijkstra__TESTED_BY__ds-ex-16', 'ds-algo-dijkstra', 'ds-ex-16', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-algo-prim__TESTED_BY__ds-ex-16', 'ds-algo-prim', 'ds-ex-16', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-seqlist__TESTED_BY__ds-ex-17', 'ds-struct-seqlist', 'ds-ex-17', 'TESTED_BY', JSON_OBJECT('kind','exercise-link')),
(@gid, 'ds-struct-linkedlist__TESTED_BY__ds-ex-17', 'ds-struct-linkedlist', 'ds-ex-17', 'TESTED_BY', JSON_OBJECT('kind','exercise-link'));

-- ==================== RELATED_TO 关系 ====================
INSERT IGNORE INTO kg_relation (graph_id, relation_id, source, target, type, properties_json)
VALUES
(@gid, 'ds-algo-dfs__RELATED_TO__ds-algo-bfs', 'ds-algo-dfs', 'ds-algo-bfs', 'RELATED_TO', JSON_OBJECT('kind','association')),
(@gid, 'ds-algo-dijkstra__RELATED_TO__ds-algo-prim', 'ds-algo-dijkstra', 'ds-algo-prim', 'RELATED_TO', JSON_OBJECT('kind','association')),
(@gid, 'ds-algo-hashfunc__RELATED_TO__ds-algo-collision', 'ds-algo-hashfunc', 'ds-algo-collision', 'RELATED_TO', JSON_OBJECT('kind','association')),
(@gid, 'ds-struct-seqlist__RELATED_TO__ds-struct-linkedlist', 'ds-struct-seqlist', 'ds-struct-linkedlist', 'RELATED_TO', JSON_OBJECT('kind','association')),
(@gid, 'ds-struct-dlist__RELATED_TO__ds-struct-clist', 'ds-struct-dlist', 'ds-struct-clist', 'RELATED_TO', JSON_OBJECT('kind','association')),
(@gid, 'ds-struct-adjmatrix__RELATED_TO__ds-struct-adjlist', 'ds-struct-adjmatrix', 'ds-struct-adjlist', 'RELATED_TO', JSON_OBJECT('kind','association'));
