package snake

// pather_test.go implements a basic world and tiles that implement Pather for
// the sake of testing.  This functionality forms the back end for
// path_test.go, and serves as an example for how to use A* for a grid.

import (
	"fmt"
	"strings"
)

// Kind* constants refer to tile kinds for input and output.
const (
	// KindPlain (.) is a plain tile with a movement cost of 1.
	KindPlain = iota
	// KindRiver (~) is a river tile with a movement cost of 2.
	KindRiver
	// KindMountain (M) is a mountain tile with a movement cost of 3.
	KindMountain
	// KindBlocker (X) is a tile which blocks movement.
	KindBlocker
	// KindFrom (F) is a tile which marks where the path should be calculated
	// from.
	KindFrom
	// KindTo (T) is a tile which marks the goal of the path.
	KindTo
	// KindPath (●) is a tile to represent where the path is in the output.
	KindPath
)

// KindRunes map tile kinds to output runes.
var KindRunes = map[int]rune{
	KindPlain:    '.',
	KindRiver:    '~',
	KindMountain: 'M',
	KindBlocker:  'X',
	KindFrom:     'F',
	KindTo:       'T',
	KindPath:     '●',
}

// RuneKinds map input runes to tile kinds.
var RuneKinds = map[rune]int{
	'.': KindPlain,
	'~': KindRiver,
	'M': KindMountain,
	'X': KindBlocker,
	'F': KindFrom,
	'T': KindTo,
}

// KindCosts map tile kinds to movement costs.
var KindCosts = map[int]float64{
	KindPlain:    1.0,
	KindFrom:     1.0,
	KindTo:       1.0,
	KindRiver:    2.0,
	KindMountain: 3.0,
}

// A Tile is a tile in a grid which implements Pather.
type Tile struct {
	// Kind is the kind of tile, potentially affecting movement.
	Kind int
	// X and Y are the coordinates of the tile.
	X, Y int
	// W is a reference to the World that the tile is a part of.
	W World
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []Pather {
	neighbors := []Pather{}
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		n := t.W.Tile(t.X+offset[0], t.Y+offset[1])
		if n != nil && n.Kind != KindBlocker {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to Pather) float64 {
	toT := to.(*Tile)
	return KindCosts[toT.Kind]
}

// PathEstimatedCost uses Manhattan distance to estimate orthogonal distance
// between non-adjacent nodes.
func (t *Tile) PathEstimatedCost(to Pather) float64 {
	toT := to.(*Tile)
	//TODO valamiért ezek közül az egyik érték elveszik néha
	if t.X == 0 || t.X == 0 {
		fmt.Println("arghhhhhhhhhhhhhhhhhhh")
		return 9999999
	}
	absX := toT.X - t.X
	if absX < 0 {
		absX = -absX
	}
	absY := toT.Y - t.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}

// World is a two dimensional map of Tiles.
type World map[int]map[int]*Tile

// Tile gets the tile at the given coordinates in the world.
func (w World) Tile(x, y int) *Tile {
	if w[x] == nil {
		return nil
	}
	return w[x][y]
}

// SetTile sets a tile at the given coordinates in the world.
func (w World) SetTile(t *Tile, x, y int) {
	if w[x] == nil {
		w[x] = map[int]*Tile{}
	}
	w[x][y] = t
	t.X = x
	t.Y = y
	t.W = w
}

// FirstOfKind gets the first tile on the board of a kind, used to get the from
// and to tiles as there should only be one of each.
func (w World) FirstOfKind(kind int) *Tile {
	for _, row := range w {
		for _, t := range row {
			if t.Kind == kind {
				return t
			}
		}
	}
	return nil
}

// From gets the from tile from the world.
func (w World) From() *Tile {
	return w.FirstOfKind(KindFrom)
}

// To gets the to tile from the world.
func (w World) To() *Tile {
	return w.FirstOfKind(KindTo)
}

// RenderPath renders a path on top of a world.
func (w World) RenderPath(path []Pather) string {
	width := len(w)
	if width == 0 {
		return ""
	}
	height := len(w[0])
	pathLocs := map[string]bool{}
	for _, p := range path {
		pT := p.(*Tile)
		pathLocs[fmt.Sprintf("%d,%d", pT.X, pT.Y)] = true
	}
	rows := make([]string, height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			t := w.Tile(x, y)
			r := ' '
			if pathLocs[fmt.Sprintf("%d,%d", x, y)] {
				r = KindRunes[KindPath]
			} else if t != nil {
				r = KindRunes[t.Kind]
			}
			rows[y] += string(r)
		}
	}
	return strings.Join(rows, "\n")
}

func (w World) getPathCoordinates(path []Pather) []Coordinate {
	pathCoordinates := make([]Coordinate, 0)
	for _, p := range path {
		pT := p.(*Tile)
		pathCoordinates = append(pathCoordinates, newCoordinate(pT.X, pT.Y))
	}
	return pathCoordinates
}

// ParseWorld parses a textual representation of a world into a world map.
func ParseWorld(input string) World {
	w := World{}
	for y, row := range strings.Split(strings.TrimSpace(input), "\n") {
		for x, raw := range row {
			kind, ok := RuneKinds[raw]
			if !ok {
				kind = KindBlocker
			}
			w.SetTile(&Tile{
				Kind: kind,
			}, x, y)
		}
	}
	return w
}

func ParseSnakeWorld(g *Game, botSnake *Snake) World {
	w := World{}
	//Board
	g.mu.Lock()
	for _, v := range g.Board.area {
		w.SetTile(&Tile{
			Kind: 0,
		}, v.x, v.y)
	}
	//Food (only for 1 food)
	for _, f := range g.Food {
		w.SetTile(&Tile{
			Kind: KindTo,
		}, f.Coordinates.x, f.Coordinates.y)
	}
	//Snakes
	for _, s := range g.Snakes {
		//snake
		for j, sp := range s.SnakeParts {

			// if j == 0 && s != botSnake {
			// 	// //test
			// 	nextCoor := nextCoorTest(s.Direction, &sp.Coordinate)
			// 	w.SetTile(&Tile{
			// 		Kind: KindBlocker,
			// 	}, nextCoor.x, nextCoor.y)
			// }
			//head
			if j == 0 && s == botSnake {
				//start
				w.SetTile(&Tile{
					Kind: KindFrom,
				}, sp.Coordinate.x, sp.Coordinate.y)

			} else {
				//Goal
				w.SetTile(&Tile{
					Kind: KindBlocker,
				}, sp.Coordinate.x, sp.Coordinate.y)
			}
		}
	}
	g.mu.Unlock()
	return w
}

func nextCoorTest(dir int, coordinates *Coordinate) *Coordinate {
	switch dir {
	case 0: //up
		coordinates.y = coordinates.y - 1
	case 1: //left
		coordinates.x = coordinates.x + 1
	case 2: //right
		coordinates.x = coordinates.x - 1
	case 3: //down
		coordinates.y = coordinates.y + 1
	}
	return coordinates
}

// g.TestField6 = fmt.Sprintf("path: %v", world.getPathCoordinates(p))
// g.TestField5 = fmt.Sprintf("dist: %v, found: %v", dist, found)
