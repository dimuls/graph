package dijkstra

import (
	"errors"

	"github.com/dimuls/graph/entity"
)

type edge struct {
	id     int64
	weight float64
	from   int64
	to     int64
}

type vertex struct {
	id       int64
	outEdges []edge
}

func initGraph(vs []entity.Vertex, es []entity.Edge) map[int64]vertex {
	vsMap := map[int64]entity.Vertex{}
	for _, v := range vs {
		vsMap[v.ID] = v
	}

	esMap := map[int64]entity.Edge{}
	for _, e := range es {
		esMap[e.ID] = e
	}

	graph := map[int64]vertex{}

	for _, v := range vs {
		v2 := vertex{
			id:       v.ID,
			outEdges: nil,
		}
		for _, e := range es {
			if e.From == v.ID {
				v2.outEdges = append(v2.outEdges, edge{
					id:     e.ID,
					weight: e.Weight,
					from:   e.From,
					to:     e.To,
				})
			}
		}
		graph[v.ID] = v2
	}

	return graph
}

func ShortestPath(vs []entity.Vertex, es []entity.Edge,
	from int64, to int64) ([]int64, error) {

	if from == to {
		return nil, nil
	}

	graph := initGraph(vs, es)

	visited := map[int64]struct{}{}
	visited[from] = struct{}{}

	distances := map[int64]float64{}
	paths := map[int64][]edge{}

	edges := graph[from].outEdges

	for {
		if len(edges) == 0 {
			break
		}

		e := edges[0]
		edges = edges[1:]

		v := graph[e.to]

		d, exists := distances[v.id]
		if !exists || d > distances[e.from]+e.weight {
			distances[v.id] = distances[e.from] + e.weight
			paths[v.id] = append(paths[e.from], e)
		}

		if _, exists := visited[v.id]; exists {
			continue
		}

		edges = append(edges, v.outEdges...)

		visited[v.id] = struct{}{}
	}

	var path []int64

	for _, e := range paths[to] {
		path = append(path, e.id)
	}

	if len(path) == 0 {
		return nil, errors.New("vertexes are not connected")
	}

	return path, nil
}
