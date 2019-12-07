package entity

type Graph struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Vertex struct {
	ID      int64   `json:"id" db:"id"`
	GraphID int64   `json:"graph_id" db:"graph_id"`
	X       float64 `json:"x" db:"x"`
	Y       float64 `json:"y" db:"y"`
}

type Edge struct {
	ID      int64   `json:"id" db:"id"`
	GraphID int64   `json:"graph_id" db:"graph_id"`
	From    int64   `json:"from" db:"from"`
	To      int64   `json:"to" db:"to"`
	Weight  float64 `json:"weight" db:"weight"`
}
