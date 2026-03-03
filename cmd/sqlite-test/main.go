package main

import (
	"compress/gzip"
	"database/sql"
	"ed-expedition/database"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gonum.org/v1/gonum/spatial/curve"
	_ "modernc.org/sqlite"
)

type dumpSystem struct {
	ID64     uint64             `json:"id64"`
	Name     string             `json:"name"`
	MainStar string             `json:"mainStar"`
	Coords   database.RawCoords `json:"coords"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := runBuild(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
			os.Exit(1)
		}
	case "index":
		if err := runIndex(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "index failed: %v\n", err)
			os.Exit(1)
		}
	case "bench":
		if err := runBench(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "bench failed: %v\n", err)
			os.Exit(1)
		}
	case "ranges":
		if err := runRanges(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "ranges failed: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("sqlite-test <build|index|bench|ranges> [flags]")
	fmt.Println("  build --input <systems.json|systems.json.gz> --db <galaxy.sqlite>")
	fmt.Println("  index --db <galaxy.sqlite> [--name] [--name8] [--name10] [--hilbert]")
	fmt.Println("  bench --db <galaxy.sqlite> [--queries 1000] [--window 0:auto] [--seed 0:time] [--prefix] [--prefix-len 8] [--prefix-plan auto|none|full|prefix8|prefix10]")
	fmt.Println("  ranges --db <galaxy.sqlite>")
}

func runRanges(args []string) error {
	fs := flag.NewFlagSet("ranges", flag.ContinueOnError)
	dbPath := fs.String("db", "", "SQLite database path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return errors.New("--db is required")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var (
		count      int64
		minX, maxX int64
		minY, maxY int64
		minZ, maxZ int64
		minH, maxH int64
	)

	err = db.QueryRow(`
		SELECT
			COUNT(*),
			MIN(x), MAX(x),
			MIN(y), MAX(y),
			MIN(z), MAX(z),
			MIN(hilbert_index), MAX(hilbert_index)
		FROM systems
	`).Scan(&count, &minX, &maxX, &minY, &maxY, &minZ, &maxZ, &minH, &maxH)
	if err != nil {
		return err
	}

	fmt.Printf("rows=%d\n", count)
	fmt.Printf("x=[%d,%d] span=%d\n", minX, maxX, maxX-minX)
	fmt.Printf("y=[%d,%d] span=%d\n", minY, maxY, maxY-minY)
	fmt.Printf("z=[%d,%d] span=%d\n", minZ, maxZ, maxZ-minZ)
	fmt.Printf("hilbert=[%d,%d] span=%d\n", minH, maxH, maxH-minH)

	return nil
}

func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	dbPath := fs.String("db", "", "SQLite database path")
	name := fs.Bool("name", true, "Create name prefix index")
	name8 := fs.Bool("name8", false, "Create name first-8-chars expression index")
	name10 := fs.Bool("name10", false, "Create name first-10-chars expression index")
	hilbert := fs.Bool("hilbert", false, "Create hilbert index")
	fast := fs.Bool("fast", true, "Use unsafe temporary PRAGMAs to speed index creation")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return errors.New("--db is required")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if *fast {
		if err := setFastIndexBuildPragmas(db); err != nil {
			return err
		}
		defer func() {
			_ = restoreDefaultPragmas(db)
		}()
	}

	if *name {
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_name ON systems(name)`); err != nil {
			return err
		}
		fmt.Println("created idx_systems_name")
	}
	if *name8 {
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_name8 ON systems(substr(name,1,8))`); err != nil {
			return err
		}
		fmt.Println("created idx_systems_name8")
	}
	if *name10 {
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_name10 ON systems(substr(name,1,10))`); err != nil {
			return err
		}
		fmt.Println("created idx_systems_name10")
	}
	if *hilbert {
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_hilbert ON systems(hilbert_index)`); err != nil {
			return err
		}
		fmt.Println("created idx_systems_hilbert")
	}

	if !*name && !*name8 && !*name10 && !*hilbert {
		fmt.Println("no indexes requested")
	}

	return nil
}

func setFastIndexBuildPragmas(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA journal_mode=OFF`,
		`PRAGMA synchronous=OFF`,
		`PRAGMA temp_store=MEMORY`,
		`PRAGMA locking_mode=EXCLUSIVE`,
		`PRAGMA cache_size=-524288`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}
	return nil
}

func restoreDefaultPragmas(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA locking_mode=NORMAL`,
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}
	return nil
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	inputPath := fs.String("input", "", "Path to systems dump (json/json.gz)")
	dbPath := fs.String("db", "", "SQLite output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *inputPath == "" || *dbPath == "" {
		return errors.New("--input and --db are required")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := prepareSchema(db); err != nil {
		return err
	}

	hilbert, err := curve.NewHilbert3D(database.HilbertOrder)
	if err != nil {
		return err
	}

	reader, closer, err := openDump(*inputPath)
	if err != nil {
		return err
	}
	defer closer()

	dec := json.NewDecoder(reader)
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return fmt.Errorf("expected JSON array start, got %v", tok)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO systems(id, hilbert_index, name, x, y, z, star_class) VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	start := time.Now()
	for dec.More() {
		var s dumpSystem
		if err := dec.Decode(&s); err != nil {
			return err
		}
		x, y, z := normalizeCoord(s.Coords.X, s.Coords.Y, s.Coords.Z)
		h := uint64(hilbert.Pos([]int{int(x), int(y), int(z)}))
		starClass := parseStarClass(s.MainStar)

		if _, err := stmt.Exec(s.ID64, h, s.Name, x, y, z, starClass); err != nil {
			return err
		}

		count++
		if count%50000 == 0 {
			if err := tx.Commit(); err != nil {
				return err
			}
			tx, err = db.Begin()
			if err != nil {
				return err
			}
			stmt, err = tx.Prepare(`INSERT INTO systems(id, hilbert_index, name, x, y, z, star_class) VALUES(?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			fmt.Printf("inserted=%d elapsed=%s\n", count, time.Since(start).Round(time.Second))
		}
	}

	if _, err := dec.Token(); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_systems_hilbert ON systems(hilbert_index)`); err != nil {
		return err
	}

	fmt.Printf("done inserted=%d total=%s\n", count, time.Since(start).Round(time.Second))
	return nil
}

func runBench(args []string) error {
	fs := flag.NewFlagSet("bench", flag.ContinueOnError)
	dbPath := fs.String("db", "", "SQLite database path")
	queries := fs.Int("queries", 1000, "Number of range queries")
	window := fs.Uint64("window", 0, "Hilbert window size (0 = auto)")
	seed := fs.Int64("seed", 0, "PRNG seed (0 = current time)")
	benchPrefix := fs.Bool("prefix", true, "Also benchmark name-prefix lookup")
	prefixLen := fs.Int("prefix-len", 8, "Prefix length for prefix benchmark")
	prefixPlan := fs.String("prefix-plan", "auto", "Prefix query plan: auto|none|full|prefix8|prefix10")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return errors.New("--db is required")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	hilbert, err := curve.NewHilbert3D(database.HilbertOrder)
	if err != nil {
		return err
	}

	if *window == 0 {
		*window = defaultHilbertWindow(hilbert)
	}
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	rng := rand.New(rand.NewSource(*seed))
	query := `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems WHERE hilbert_index BETWEEN ? AND ?`
	starClassFilteredQuery := `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems WHERE hilbert_index BETWEEN ? AND ? AND star_class IN (?, ?, ?, ?, ?, ?, ?)`

	start := time.Now()
	totalRows := 0
	prefixes := make([]string, 0, *queries)
	for i := 0; i < *queries; i++ {
		centerKey := randomHilbertKey(rng, hilbert)
		from := centerKey - *window
		to := centerKey + *window

		rows, err := db.Query(query, from, to)
		if err != nil {
			return err
		}

		for rows.Next() {
			var (
				id      uint64
				h       uint64
				name    string
				x       uint32
				y       uint32
				z       uint32
				starCls uint8
			)
			if err := rows.Scan(&id, &h, &name, &x, &y, &z, &starCls); err != nil {
				rows.Close()
				return err
			}
			if *benchPrefix && len(prefixes) < i+1 {
				prefix := name
				if len(prefix) > *prefixLen {
					prefix = prefix[:*prefixLen]
				}
				if len(prefix) > 0 {
					prefixes = append(prefixes, prefix)
				}
			}
			totalRows++
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}

	elapsed := time.Since(start)
	fmt.Printf("queries=%d seed=%d window=%d total=%s avg=%s rows=%d\n", *queries, *seed, *window, elapsed.Round(time.Millisecond), (elapsed / time.Duration(*queries)).Round(time.Microsecond), totalRows)

	filteredStart := time.Now()
	filteredTotalRows := 0
	for i := 0; i < *queries; i++ {
		centerKey := randomHilbertKey(rng, hilbert)
		from := centerKey - *window
		to := centerKey + *window

		rows, err := db.Query(
			starClassFilteredQuery,
			from,
			to,
			highSequenceStarClasses[0],
			highSequenceStarClasses[1],
			highSequenceStarClasses[2],
			highSequenceStarClasses[3],
			highSequenceStarClasses[4],
			highSequenceStarClasses[5],
			highSequenceStarClasses[6],
		)
		if err != nil {
			return err
		}

		for rows.Next() {
			var (
				id      uint64
				h       uint64
				name    string
				x       uint32
				y       uint32
				z       uint32
				starCls uint8
			)
			if err := rows.Scan(&id, &h, &name, &x, &y, &z, &starCls); err != nil {
				rows.Close()
				return err
			}
			filteredTotalRows++
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}

	filteredElapsed := time.Since(filteredStart)
	fmt.Printf("filtered_queries=%d seed=%d window=%d classes=7 total=%s avg=%s rows=%d\n", *queries, *seed, *window, filteredElapsed.Round(time.Millisecond), (filteredElapsed / time.Duration(*queries)).Round(time.Microsecond), filteredTotalRows)

	if *benchPrefix && len(prefixes) > 0 {
		prefixQuery := `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems WHERE name >= ? AND name < ? LIMIT 10`
		prefixUsesName10 := false
		switch *prefixPlan {
		case "auto":
		case "none":
			prefixQuery = `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems NOT INDEXED WHERE name >= ? AND name < ? LIMIT 10`
		case "full":
			prefixQuery = `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems INDEXED BY idx_systems_name WHERE name >= ? AND name < ? LIMIT 10`
		case "prefix8":
			prefixQuery = `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems INDEXED BY idx_systems_name8 WHERE substr(name,1,8) >= ? AND substr(name,1,8) < ? AND name >= ? AND name < ? LIMIT 10`
			prefixUsesName10 = true
		case "prefix10":
			prefixQuery = `SELECT id, hilbert_index, name, x, y, z, star_class FROM systems INDEXED BY idx_systems_name10 WHERE substr(name,1,10) >= ? AND substr(name,1,10) < ? AND name >= ? AND name < ? LIMIT 10`
			prefixUsesName10 = true
		default:
			return fmt.Errorf("invalid --prefix-plan %q (expected auto|none|full|prefix8|prefix10)", *prefixPlan)
		}
		prefixStart := time.Now()
		prefixRowsTotal := 0
		prefixQueriesRan := 0

		for _, prefix := range prefixes {
			upper, ok := nextPrefix(prefix)
			if !ok {
				continue
			}
			var rows *sql.Rows
			if prefixUsesName10 {
				prefixWidth := 10
				if *prefixPlan == "prefix8" {
					prefixWidth = 8
				}
				from10, to10, ok := prefixNBounds(prefix, prefixWidth)
				if !ok {
					continue
				}
				rows, err = db.Query(prefixQuery, from10, to10, prefix, upper)
			} else {
				rows, err = db.Query(prefixQuery, prefix, upper)
			}
			if err != nil {
				return err
			}
			prefixQueriesRan++
			for rows.Next() {
				var (
					id      uint64
					h       uint64
					name    string
					x       uint32
					y       uint32
					z       uint32
					starCls uint8
				)
				if err := rows.Scan(&id, &h, &name, &x, &y, &z, &starCls); err != nil {
					rows.Close()
					return err
				}
				prefixRowsTotal++
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return err
			}
			rows.Close()
		}

		prefixElapsed := time.Since(prefixStart)
		if prefixQueriesRan > 0 {
			fmt.Printf("prefix_queries=%d plan=%s prefix_len=%d total=%s avg=%s rows=%d\n", prefixQueriesRan, *prefixPlan, *prefixLen, prefixElapsed.Round(time.Millisecond), (prefixElapsed / time.Duration(prefixQueriesRan)).Round(time.Microsecond), prefixRowsTotal)
		} else {
			fmt.Printf("prefix_queries=0 plan=%s prefix_len=%d total=%s avg=0s rows=0\n", *prefixPlan, *prefixLen, prefixElapsed.Round(time.Millisecond))
		}

	}
	return nil
}

func nextPrefix(s string) (string, bool) {
	b := []byte(s)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] < 0xFF {
			b[i]++
			return string(b[:i+1]), true
		}
	}
	return "", false
}

func prefixNBounds(prefix string, n int) (string, string, bool) {
	p := prefix
	if len(p) > n {
		p = p[:n]
	}
	upper, ok := nextPrefix(p)
	if !ok {
		return "", "", false
	}
	return p, upper, true
}

func defaultHilbertWindow(h curve.Hilbert3D) uint64 {
	points := [][3]float64{
		{0, 0, 0},
		{10, 0, 0},
		{0, 12, 0},
		{0, 0, 15},
	}

	minKey := uint64(^uint64(0))
	maxKey := uint64(0)
	for _, p := range points {
		x, y, z := normalizeCoord(p[0], p[1], p[2])
		k := uint64(h.Pos([]int{int(x), int(y), int(z)}))
		if k < minKey {
			minKey = k
		}
		if k > maxKey {
			maxKey = k
		}
	}

	if maxKey <= minKey {
		return 50000
	}
	return maxKey - minKey
}

func randomHilbertKey(rng *rand.Rand, h curve.Hilbert3D) uint64 {
	// Sample around the center of observed populated bounds using 50% of each axis span.
	const (
		minX = 7861.0
		maxX = 835038.0
		minY = 6401.0
		maxY = 695183.0
		minZ = 5950.0
		maxZ = 896301.0
	)

	spanX := maxX - minX
	spanY := maxY - minY
	spanZ := maxZ - minZ

	centerX := minX + (spanX / 2)
	centerY := minY + (spanY / 2)
	centerZ := minZ + (spanZ / 2)

	halfRangeX := spanX * 0.25
	halfRangeY := spanY * 0.25
	halfRangeZ := spanZ * 0.25

	nx := centerX + (rng.Float64()*2-1)*halfRangeX
	ny := centerY + (rng.Float64()*2-1)*halfRangeY
	nz := centerZ + (rng.Float64()*2-1)*halfRangeZ

	if nx < minX {
		nx = minX
	}
	if nx > maxX {
		nx = maxX
	}
	if ny < minY {
		ny = minY
	}
	if ny > maxY {
		ny = maxY
	}
	if nz < minZ {
		nz = minZ
	}
	if nz > maxZ {
		nz = maxZ
	}

	return uint64(h.Pos([]int{int(nx), int(ny), int(nz)}))
}

func prepareSchema(db *sql.DB) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
		`DROP TABLE IF EXISTS systems`,
		`CREATE TABLE systems (
			id INTEGER PRIMARY KEY,
			hilbert_index INTEGER NOT NULL,
			name TEXT NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			z INTEGER NOT NULL,
			star_class INTEGER NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func openDump(path string) (io.Reader, func() error, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	closeFn := func() error { return f.Close() }
	if filepath.Ext(path) == ".gz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		closeFn = func() error {
			e1 := gz.Close()
			e2 := f.Close()
			if e1 != nil {
				return e1
			}
			return e2
		}
		return gz, closeFn, nil
	}

	return f, closeFn, nil
}

func normalizeCoord(x, y, z float64) (nx uint32, ny uint32, nz uint32) {
	nx = uint32((x - database.OriginX) * database.CoordScale)
	ny = uint32((y - database.OriginY) * database.CoordScale)
	nz = uint32((z - database.OriginZ) * database.CoordScale)
	return nx, ny, nz
}

var starClassMap = map[string]uint8{
	"O (Blue-White) Star":               0x01,
	"B (Blue-White) Star":               0x02,
	"A (Blue-White) Star":               0x03,
	"F (White) Star":                    0x04,
	"G (White-Yellow) Star":             0x05,
	"K (Yellow-Orange) Star":            0x06,
	"M (Red dwarf) Star":                0x07,
	"K (Yellow-Orange giant) Star":      0x10,
	"M (Red giant) Star":                0x11,
	"M (Red super giant) Star":          0x12,
	"A (Blue-White super giant) Star":   0x13,
	"B (Blue-White super giant) Star":   0x14,
	"F (White super giant) Star":        0x15,
	"G (White-Yellow super giant) Star": 0x16,
	"L (Brown dwarf) Star":              0x20,
	"T (Brown dwarf) Star":              0x21,
	"Y (Brown dwarf) Star":              0x22,
	"C Star":                            0x30,
	"CN Star":                           0x31,
	"CJ Star":                           0x32,
	"MS-type Star":                      0x33,
	"S-type Star":                       0x34,
	"White Dwarf (D) Star":              0x40,
	"White Dwarf (DA) Star":             0x41,
	"White Dwarf (DAB) Star":            0x42,
	"White Dwarf (DAV) Star":            0x43,
	"White Dwarf (DAZ) Star":            0x44,
	"White Dwarf (DB) Star":             0x45,
	"White Dwarf (DBV) Star":            0x46,
	"White Dwarf (DBZ) Star":            0x47,
	"White Dwarf (DC) Star":             0x48,
	"White Dwarf (DCV) Star":            0x49,
	"White Dwarf (DQ) Star":             0x4A,
	"Wolf-Rayet Star":                   0x60,
	"Wolf-Rayet C Star":                 0x61,
	"Wolf-Rayet N Star":                 0x62,
	"Wolf-Rayet NC Star":                0x63,
	"Wolf-Rayet O Star":                 0x64,
	"T Tauri Star":                      0x70,
	"Herbig Ae/Be Star":                 0x71,
	"Neutron Star":                      0x80,
	"Black Hole":                        0x81,
	"Supermassive Black Hole":           0x82,
}

var highSequenceStarClasses = [7]uint8{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

func parseStarClass(starType string) uint8 {
	if starType == "" {
		return 0x00
	}
	if class, ok := starClassMap[starType]; ok {
		return class
	}
	if strings.TrimSpace(starType) == "" {
		return 0x00
	}
	return 0x00
}
