// internal/app/core/scheduler/sorter.go
package scheduler

import (
	"fmt"

	"github.com/MadeByDoug/glint/internal/app/infra/config"
)

// SortRules performs a topological sort on the given rules to determine a valid
// execution order. It returns the sorted list of rules or an error if a
// circular dependency is detected.
func SortRules(rules []config.RuleConfig) ([]config.RuleConfig, error) {
	// --- 1. Build the dependency graph and in-degree counts ---
	graph := make(map[string][]string)    // Key: dependency ID, Value: IDs of rules that depend on it
	inDegree := make(map[string]int)      // Key: rule ID, Value: count of its dependencies
	ruleMap := make(map[string]config.RuleConfig) // Key: rule ID, Value: the rule object

	for _, rule := range rules {
		if _, exists := ruleMap[rule.ID]; exists {
			return nil, fmt.Errorf("duplicate rule ID found: %s", rule.ID)
		}
		ruleMap[rule.ID] = rule
		inDegree[rule.ID] = 0
		graph[rule.ID] = []string{}
	}

	for _, rule := range rules {
		for _, depID := range rule.DependsOn {
			if _, exists := ruleMap[depID]; !exists {
				return nil, fmt.Errorf("rule '%s' depends on non-existent rule '%s'", rule.ID, depID)
			}
			// Edge: depID -> rule.ID
			graph[depID] = append(graph[depID], rule.ID)
			inDegree[rule.ID]++
		}
	}

	// --- 2. Initialize the queue with rules that have no dependencies ---
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// --- 3. Process the queue to build the sorted list ---
	var sortedRules []config.RuleConfig
	for len(queue) > 0 {
		// Dequeue
		currentID := queue[0]
		queue = queue[1:]

		sortedRules = append(sortedRules, ruleMap[currentID])

		// For each rule that depended on the one we just processed...
		for _, dependentID := range graph[currentID] {
			inDegree[dependentID]--
			// ...if it now has no other unmet dependencies, add it to the queue.
			if inDegree[dependentID] == 0 {
				queue = append(queue, dependentID)
			}
		}
	}

	// --- 4. Check for circular dependencies ---
	if len(sortedRules) != len(rules) {
		return nil, fmt.Errorf(
			"circular dependency detected; %d rules sorted out of %d",
			len(sortedRules),
			len(rules),
		)
	}

	return sortedRules, nil
}
