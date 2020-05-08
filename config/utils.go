package config

import (
	"fmt"
	"strings"

	C "github.com/ClashrAuto/Clashr/constant"
	adapters "github.com/ClashrAuto/Clashr/adapters/outbound"
	"github.com/ClashrAuto/Clashr/common/structure"
)

func trimArr(arr []string) (r []string) {
	for _, e := range arr {
		r = append(r, strings.Trim(e, " "))
	}
	return
}

func getProxies(mapping map[string]C.Proxy, list []string) ([]C.Proxy, error) {
	var ps []C.Proxy
	for _, name := range list {
		p, ok := mapping[name]
		if !ok {
			return nil, fmt.Errorf("'%s' not found", name)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func or(pointers ...*int) *int {
	for _, p := range pointers {
		if p != nil {
			return p
		}
	}
	return pointers[len(pointers)-1]
}

// Check if ProxyGroups form DAG(Directed Acyclic Graph), and sort all ProxyGroups by dependency order.
// Meanwhile, record the original index in the config file.
// If loop is detected, return an error with location of loop.
func proxyGroupsDagSort(groupsConfig []map[string]interface{}, decoder *structure.Decoder) error {

	type graphNode struct {
		indegree int
		// topological order
		topo int
		// the origional data in `groupsConfig`
		data map[string]interface{}
		// `outdegree` and `from` are used in loop locating
		outdegree int
		from      []string
	}

	graph := make(map[string]*graphNode)

	// Step 1.1 build dependency graph
	for _, mapping := range groupsConfig {
		option := &adapters.ProxyGroupOption{}
		err := decoder.Decode(mapping, option)
		groupName := option.Name
		if err != nil {
			return fmt.Errorf("ProxyGroup %s: %s", groupName, err.Error())
		}
		if node, ok := graph[groupName]; ok {
			if node.data != nil {
				return fmt.Errorf("ProxyGroup %s: duplicate group name", groupName)
			}
			node.data = mapping
		} else {
			graph[groupName] = &graphNode{0, -1, mapping, 0, nil}
		}

		for _, proxy := range option.Proxies {
			if node, ex := graph[proxy]; ex {
				node.indegree++
			} else {
				graph[proxy] = &graphNode{1, -1, nil, 0, nil}
			}
		}
	}
	// Step 1.2 Topological Sort
	// topological index of **ProxyGroup**
	index := 0
	queue := make([]string, 0)
	for name, node := range graph {
		// in the begning, put nodes that have `node.indegree == 0` into queue.
		if node.indegree == 0 {
			queue = append(queue, name)
		}
	}
	// every element in queue have indegree == 0
	for ; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		node := graph[name]
		if node.data != nil {
			index++
			groupsConfig[len(groupsConfig)-index] = node.data
			for _, proxy := range node.data["proxies"].([]interface{}) {
				child := graph[proxy.(string)]
				child.indegree--
				if child.indegree == 0 {
					queue = append(queue, proxy.(string))
				}
			}
		}
		delete(graph, name)
	}

	// no loop is detected, return sorted ProxyGroup
	if len(graph) == 0 {
		return nil
	}

	// if loop is detected, locate the loop and throw an error
	// Step 2.1 rebuild the graph, fill `outdegree` and `from` filed
	for name, node := range graph {
		if node.data == nil {
			continue
		}
		for _, proxy := range node.data["proxies"].([]interface{}) {
			node.outdegree++
			child := graph[proxy.(string)]
			if child.from == nil {
				child.from = make([]string, 0, child.indegree)
			}
			child.from = append(child.from, name)
		}
	}
	// Step 2.2 remove nodes outside the loop. so that we have only the loops remain in `graph`
	queue = make([]string, 0)
	// initialize queue with node have outdegree == 0
	for name, node := range graph {
		if node.outdegree == 0 {
			queue = append(queue, name)
		}
	}
	// every element in queue have outdegree == 0
	for ; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		node := graph[name]
		for _, f := range node.from {
			graph[f].outdegree--
			if graph[f].outdegree == 0 {
				queue = append(queue, f)
			}
		}
		delete(graph, name)
	}
	// Step 2.3 report the elements in loop
	loopElements := make([]string, 0, len(graph))
	for name := range graph {
		loopElements = append(loopElements, name)
		delete(graph, name)
	}
	return fmt.Errorf("Loop is detected in ProxyGroup, please check following ProxyGroups: %v", loopElements)
}
