package database

import (
	"database/sql"
	"ed-expedition/lib/vec"
	"fmt"
	"path/filepath"
	"strings"

	"gonum.org/v1/gonum/spatial/curve"

	_ "modernc.org/sqlite"
)

const (
	// Hilbert curve: order 20 for ~0.1 ly precision (fits in 60 bits)
	HilbertOrder = 20
	HilbertBits  = 60

	CoordScale float64 = 10 // 0.1 ly precision
)

var (
	hilbert curve.Hilbert3D
	Origin  = vec.Vec3{X: -43000, Y: -30000, Z: -24000}
)

func init() {
	var err error
	hilbert, err = curve.NewHilbert3D(HilbertOrder)
	if err != nil {
		panic(fmt.Sprintf("Could not create hilbert curve: %s", err.Error()))
	}
}

func Hilbert(x, y, z uint32) int {
	return hilbert.Pos([]int{int(x), int(y), int(z)})
}

func GalaxyDBPath() string {
	return filepath.Join(DataDir, "galaxy.sqlite")
}

type queryable interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type galaxyQuerier struct {
	q queryable
}

func (g *galaxyQuerier) PrepareSystemInsert() (*sql.Stmt, error) {
	return g.q.Prepare(
		`INSERT INTO systems
			(id, hilbert_index, name, x, y, z, star_class)
			VALUES(?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO NOTHING`,
	)
}

func (g *galaxyQuerier) EnsureSystemsTable() error {
	_, err := g.q.Exec(`
		CREATE TABLE IF NOT EXISTS systems (
			id INTEGER PRIMARY KEY,
			hilbert_index INTEGER NOT NULL,
			name TEXT NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			z INTEGER NOT NULL,
			star_class INTEGER NOT NULL
		)
	`)
	return err
}

func (g *galaxyQuerier) EnsureSystemsIndexes() error {
	if _, err := g.q.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_hilbert ON systems(hilbert_index)`); err != nil {
		return err
	}
	if _, err := g.q.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_name ON systems(name COLLATE NOCASE)`); err != nil {
		return err
	}
	return nil
}

func (g *galaxyQuerier) SystemsByPrefix(prefix string, limit int) ([]*System, error) {
	rows, err := g.q.Query(
		`SELECT id, hilbert_index, name, x, y, z, star_class
		FROM systems
		WHERE name LIKE ? COLLATE NOCASE
		LIMIT ?`,
		prefix+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]*System, 0, limit)
	for rows.Next() {
		var sys System
		if err := rows.Scan(&sys.Id, &sys.hilbertKey, &sys.Name, &sys.X, &sys.Y, &sys.Z, &sys.StarClass); err != nil {
			return nil, err
		}
		systems = append(systems, &sys)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return systems, nil
}

func (g *galaxyQuerier) SystemByName(name string) (*System, error) {
	row := g.q.QueryRow(
		`SELECT id, hilbert_index, name, x, y, z, star_class 
		FROM systems 
		WHERE name = ? COLLATE NOCASE 
		LIMIT 1`,
		name,
	)

	var sys System
	err := row.Scan(&sys.Id, &sys.hilbertKey, &sys.Name, &sys.X, &sys.Y, &sys.Z, &sys.StarClass)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &sys, nil
}

func (g *galaxyQuerier) SystemsByHilbertRange(min, max int) ([]*System, error) {
	rows, err := g.q.Query(
		`SELECT id, hilbert_index, name, x, y, z, star_class 
		FROM systems 
		WHERE hilbert_index BETWEEN ? AND ?`,
		min, max,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]*System, 0, 128)
	for rows.Next() {
		var sys System
		if err := rows.Scan(&sys.Id, &sys.hilbertKey, &sys.Name, &sys.X, &sys.Y, &sys.Z, &sys.StarClass); err != nil {
			return nil, err
		}
		systems = append(systems, &sys)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return systems, nil
}

func (g *galaxyQuerier) SystemsByHilbertRanges(ranges [][2]int) ([]*System, error) {
	if len(ranges) == 0 {
		return []*System{}, fmt.Errorf("You need to provide at least one range")
	}

	clauses := make([]string, 0, len(ranges))
	for _, r := range ranges {
		clauses = append(clauses, fmt.Sprintf("hilbert_index BETWEEN %d AND %d", r[0], r[1]))
	}

	query := fmt.Sprintf(
		`SELECT id, hilbert_index, name, x, y, z, star_class 
		FROM systems 
		WHERE %s`,
		strings.Join(clauses, " OR "),
	)

	rows, err := g.q.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]*System, 0, 128)
	for rows.Next() {
		var sys System
		if err := rows.Scan(&sys.Id, &sys.hilbertKey, &sys.Name, &sys.X, &sys.Y, &sys.Z, &sys.StarClass); err != nil {
			return nil, err
		}
		systems = append(systems, &sys)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return systems, nil
}

func (g *galaxyQuerier) ListTables() ([]string, error) {
	rows, err := g.q.Query(`SELECT name FROM sqlite_master WHERE type = 'table'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (g *galaxyQuerier) ListIndexesForTable(tableName string) ([]string, error) {
	rows, err := g.q.Query(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ?`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		indexes = append(indexes, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

type GalaxyDB struct {
	*sql.DB
	galaxyQuerier
}

func OpenGalaxyDB() (*GalaxyDB, error) {
	db, err := sql.Open("sqlite", GalaxyDBPath())
	if err != nil {
		return nil, err
	}

	return &GalaxyDB{
		DB:            db,
		galaxyQuerier: galaxyQuerier{q: db},
	}, nil
}

func (g *GalaxyDB) Begin() (*GalaxyTx, error) {
	tx, err := g.DB.Begin()
	if err != nil {
		return nil, err
	}

	return &GalaxyTx{
		Tx:            tx,
		galaxyQuerier: galaxyQuerier{q: tx},
	}, nil
}

type GalaxyTx struct {
	*sql.Tx
	galaxyQuerier
}
