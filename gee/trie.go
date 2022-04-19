package gee

import "strings"

type node struct {
	pattern string // 待匹配路由，例如 /p/:lang
	// 只有叶子节点会记录 pattern，非叶子节点 pattern 为 ""
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否模糊匹配，part 含有 : 或 * 时为 true
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

/** 插入新节点（注册新路由）
 * pattern: /p/:lang/doc
 * parts: [p, :lang, doc]
 * height: len(parts) -> 3
 * n: 当前层的节点，第 0 层为 '/' 节点，每次在它子节点中找 part
 */
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	// 如果没有匹配到当前层的 part 的节点，则新建一个
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

/** 查找路由
 *  返回叶子节点
 */
func (n *node) search(parts []string, height int) *node {
	// 如果检查到当前 pattern 的最后一层，则两种情况下匹配成功
	// 1. 当前 pattern 的最后一层是某个叶子节点
	// 2. 当前 pattern 的最后一层对应的某层节点为通配节点
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" { // 说明是非叶子节点
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		// 只要有一条正确路径即可
		if result != nil {
			return result
		}
	}

	return nil
}
