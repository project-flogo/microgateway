package graphql

import (
	"fmt"
	"math"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// calculateQueryDepth calculates graphQL request query depth
func calculateQueryDepth(requestString string) int {
	depth := 0

	source := source.NewSource(&source.Source{
		Body: []byte(requestString),
		Name: "GraphQL request",
	})

	// parse the source
	AST, err := parser.Parse(parser.ParseParams{Source: source})
	if err != nil {
		fmt.Println("parse error: ", err)
		return -1
	}

	// get queries, mutations & fragments
	queriesOrMutations := make(map[string]interface{})
	fragments := make(map[string]interface{})
	for i, node := range AST.Definitions {
		if node.GetKind() == kinds.OperationDefinition {
			opDef := node.(*ast.OperationDefinition)
			if name := opDef.GetName(); name != nil {
				queriesOrMutations[name.Value] = node
			} else {
				queriesOrMutations[string(i)] = node
			}
		} else if node.GetKind() == kinds.FragmentDefinition {
			fDef := node.(*ast.FragmentDefinition)
			if name := fDef.GetName(); name != nil {
				fragments[name.Value] = node
			} else {
				fragments[string(i)] = node
			}
		}
	}
	depths := make([]int, len(queriesOrMutations))

	// calculate queryDepth for all queries/mutations
	i := 0
	for _, q := range queriesOrMutations {
		d, err := calculateDepth(q, fragments, 0, math.MaxInt8)
		if err != nil {
			fmt.Println("error: ", err)
		}
		depths[i] = d
		i++
	}

	// get max depth from depths
	for _, d := range depths {
		if depth < d {
			depth = d
		}
	}

	return depth
}

func calculateDepth(node interface{}, fragments map[string]interface{}, depthSoFar int, maxDepth int) (int, error) {
	if depthSoFar > maxDepth {
		errMsg := fmt.Sprintf("exceeds maximum allowed depth[%v]", maxDepth)
		return -1, fmt.Errorf(errMsg)
	}

	switch v := node.(type) {

	case *ast.Document:
		depth := 0
		for _, n := range v.Definitions {
			d, err := calculateDepth(n, fragments, depthSoFar, maxDepth)
			if err != nil {
				return -1, err
			}
			if depth < d {
				depth = d
			}
		}
		return depth, nil

	case *ast.OperationDefinition:
		return calculateDepth(v.SelectionSet, fragments, depthSoFar, maxDepth)

	case *ast.FragmentDefinition:
		return calculateDepth(v.SelectionSet, fragments, depthSoFar, maxDepth)

	case *ast.InlineFragment:
		return calculateDepth(v.SelectionSet, fragments, depthSoFar, maxDepth)

	case *ast.SelectionSet:
		depth := 0
		for _, n := range v.Selections {
			d, err := calculateDepth(n, fragments, depthSoFar, maxDepth)
			if err != nil {
				return -1, err
			}
			if depth < d {
				depth = d
			}
		}
		return depth, nil

	case *ast.Field:
		if v.SelectionSet == nil {
			if name := v.Name; name != nil {
				fmt.Printf("%s[depth=%v] \n", name.Value, depthSoFar+1)
			}
			return 1, nil
		}
		depth, err := calculateDepth(v.GetSelectionSet(), fragments, depthSoFar+1, maxDepth)
		return depth + 1, err

	case *ast.FragmentSpread:
		return calculateDepth(fragments[v.Name.Value], fragments, depthSoFar, maxDepth)

	default:
		return -1, fmt.Errorf("cannot handle node type[%T]", node)
	}
}
