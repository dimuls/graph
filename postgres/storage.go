package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/driver/postgres"
	"github.com/dimuls/graph/entity"
	"github.com/gobuffalo/packr"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db  *sqlx.DB
	uri string
}

func NewStorage(postgresURI string) (*Storage, error) {
	db, err := sqlx.Open("postgres", postgresURI)
	if err != nil {
		return nil, errors.New("failed to open DB: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)

	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		return nil, errors.New("failed to ping DB: " + err.Error())
	}

	return &Storage{db: db, uri: postgresURI}, nil
}

//go:generate packr

const migrationsPath = "./migrations"

func (s *Storage) Migrate() error {
	packrSource := &migration.PackrMigrationSource{
		Box: packr.NewBox(migrationsPath),
	}

	d, err := postgres.New(s.uri)
	if err != nil {
		return errors.New("failed to create migration driver: " + err.Error())
	}

	_, err = migration.Migrate(d, packrSource, migration.Up, 0)
	if err != nil {
		return errors.New("failed to migrate: " + err.Error())
	}

	return nil
}

func (s *Storage) Graph(graphID int64) (g entity.Graph, err error) {
	err = s.db.QueryRowx(`SELECT * FROM graph WHERE id = $1`, graphID).
		StructScan(&g)
	if err == sql.ErrNoRows {
		err = entity.ErrGraphNotFound
	}
	return
}

func (s *Storage) Graphs() (gs []entity.Graph, err error) {
	err = s.db.Select(&gs, `SELECT * FROM graph ORDER BY name`)
	return
}

func (s *Storage) AddGraph(g entity.Graph) (id int64, err error) {
	err = s.db.QueryRow(`
		INSERT INTO graph (name) VALUES ($1) RETURNING id
	`, g.Name).Scan(&id)
	if terr, ok := err.(*pq.Error); ok {
		if terr.Code == "23505" { // duplicate key violates unique constraint
			err = entity.ErrDuplicatedGraphName
		}
	}
	return
}

func (s *Storage) RemoveGraph(graphID int64) (err error) {
	_, err = s.db.Exec(`DELETE FROM graph WHERE id = $1`, graphID)
	return
}

func (s *Storage) Vertex(vertexID int64) (v entity.Vertex, err error) {
	err = s.db.QueryRowx(`SELECT * FROM vertex WHERE id = $1`,
		vertexID).StructScan(&v)
	if err == sql.ErrNoRows {
		err = entity.ErrVertexNotFound
	}
	return
}

func (s *Storage) Vertexes(graphID int64) (vs []entity.Vertex, err error) {
	err = s.db.Select(&vs, `SELECT * FROM vertex WHERE graph_id = $1`,
		graphID)
	return
}

func (s *Storage) AddVertex(v entity.Vertex) (id int64, err error) {
	err = s.db.QueryRow(`
		INSERT INTO vertex (graph_id, x, y)
		VALUES ($1, $2, $3)
		RETURNING id
	`, v.GraphID, v.X, v.Y).Scan(&id)
	return
}

func (s *Storage) SetVertex(v entity.Vertex) error {
	res, err := s.db.Exec(`
		UPDATE vertex SET x = $1, y = $2 WHERE id = $3
	`, v.X, v.Y, v.ID)
	if err != nil {
		return err
	}
	updates, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if updates == 0 {
		return entity.ErrVertexNotFound
	}
	return nil
}

func (s *Storage) RemoveVertex(vertexID int64) (err error) {
	_, err = s.db.Exec(`DELETE FROM vertex WHERE id = $1`, vertexID)
	return
}

func (s *Storage) Edge(edgeID int64) (e entity.Edge, err error) {
	err = s.db.QueryRowx(`SELECT * FROM edge WHERE id = $1`,
		edgeID).StructScan(&e)
	if err == sql.ErrNoRows {
		err = entity.ErrEdgeNotFound
	}
	return
}

func (s *Storage) Edges(graphID int64) (es []entity.Edge, err error) {
	err = s.db.Select(&es, `SELECT * FROM edge WHERE graph_id = $1`,
		graphID)
	return
}

func (s *Storage) AddEdge(e entity.Edge) (id int64, err error) {
	err = s.db.QueryRow(`
		INSERT INTO edge (graph_id, "from", "to", weight)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, e.GraphID, e.From, e.To, e.Weight).Scan(&id)
	return
}

func (s *Storage) SetEdge(e entity.Edge) error {
	res, err := s.db.Exec(`
		UPDATE edge SET weight = $1 WHERE id = $2
	`, e.Weight, e.ID)
	if err != nil {
		return err
	}
	updates, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if updates == 0 {
		return entity.ErrEdgeNotFound
	}
	return nil
}

func (s *Storage) RemoveEdge(edgeID int64) (err error) {
	_, err = s.db.Exec(`DELETE FROM edge WHERE id = $1`, edgeID)
	return
}
