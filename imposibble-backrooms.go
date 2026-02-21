package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

///////////////////////////////////////////////////////////////
// GLOBAL VARIABLES
///////////////////////////////////////////////////////////////

// Maze size grows every level
var currentSize = 31

// Dynamic 2D maze map
var gameMap [][]int

///////////////////////////////////////////////////////////////
// PLAYER STRUCT
///////////////////////////////////////////////////////////////

type Player struct {
	x, y  float64
	angle float64
	speed float64
}

///////////////////////////////////////////////////////////////
// GAME STRUCT
///////////////////////////////////////////////////////////////

type Game struct {
	player   Player
	level    int
	finished bool
}

///////////////////////////////////////////////////////////////
// MAZE GENERATION (DFS + CHAOS LOOPS)
///////////////////////////////////////////////////////////////

func generateMaze() {

	// Create dynamic 2D slice
	gameMap = make([][]int, currentSize)
	for i := range gameMap {
		gameMap[i] = make([]int, currentSize)
	}

	// Fill everything with walls
	for y := 0; y < currentSize; y++ {
		for x := 0; x < currentSize; x++ {
			gameMap[y][x] = 1
		}
	}

	type cell struct{ x, y int }
	stack := []cell{{1, 1}}
	gameMap[1][1] = 0

	rand.Seed(time.Now().UnixNano())
	dirs := []cell{{0, -2}, {2, 0}, {0, 2}, {-2, 0}}

	// Standard DFS maze generation
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		rand.Shuffle(len(dirs), func(i, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})

		for _, d := range dirs {
			nx := current.x + d.x
			ny := current.y + d.y

			if nx > 0 && nx < currentSize-1 &&
				ny > 0 && ny < currentSize-1 &&
				gameMap[ny][nx] == 1 {

				gameMap[ny][nx] = 0
				gameMap[current.y+d.y/2][current.x+d.x/2] = 0
				stack = append(stack, cell{nx, ny})
			}
		}
	}

	///////////////////////////////////////////////////////////
	// CHAOS PHASE
	// Break random walls to create loops (wall-following fails)
	///////////////////////////////////////////////////////////

	for i := 0; i < currentSize*currentSize/4; i++ {

		x := rand.Intn(currentSize-2) + 1
		y := rand.Intn(currentSize-2) + 1

		if gameMap[y][x] == 1 {

			openCount := 0

			if gameMap[y+1][x] == 0 {
				openCount++
			}
			if gameMap[y-1][x] == 0 {
				openCount++
			}
			if gameMap[y][x+1] == 0 {
				openCount++
			}
			if gameMap[y][x-1] == 0 {
				openCount++
			}

			if openCount >= 2 {
				gameMap[y][x] = 0
			}
		}
	}
}

///////////////////////////////////////////////////////////////
// PLACE EXIT AT FARTHEST POINT
///////////////////////////////////////////////////////////////

func placeExit(px, py int) {

	var maxDist float64
	var exitX, exitY int

	for y := 1; y < currentSize-1; y++ {
		for x := 1; x < currentSize-1; x++ {
			if gameMap[y][x] == 0 {
				d := math.Hypot(float64(px-x), float64(py-y))
				if d > maxDist {
					maxDist = d
					exitX, exitY = x, y
				}
			}
		}
	}

	gameMap[exitY][exitX] = 2
}

///////////////////////////////////////////////////////////////
// UPDATE
///////////////////////////////////////////////////////////////

func (g *Game) Update() error {

	// FINAL SCREEN STATE
	if g.finished {
		if ebiten.IsKeyPressed(ebiten.KeyR) {

			// Reset everything
			g.level = 1
			g.finished = false
			currentSize = 31

			generateMaze()
			placeExit(1, 1)

			g.player.x = 1.5
			g.player.y = 1.5
			g.player.angle = 0
		}
		return nil
	}

	p := &g.player

	// Movement
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		nx := p.x + math.Cos(p.angle)*p.speed
		ny := p.y + math.Sin(p.angle)*p.speed
		if gameMap[int(ny)][int(nx)] != 1 {
			p.x = nx
			p.y = ny
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		nx := p.x - math.Cos(p.angle)*p.speed
		ny := p.y - math.Sin(p.angle)*p.speed
		if gameMap[int(ny)][int(nx)] != 1 {
			p.x = nx
			p.y = ny
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.angle -= 0.05
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.angle += 0.05
	}

	// Check exit
	if gameMap[int(p.y)][int(p.x)] == 2 {

		// Final level reached
		if g.level == 4 {
			g.finished = true
			return nil
		}

		// Increase difficulty
		g.level++

		// Grow maze
		currentSize += 10
		if currentSize%2 == 0 {
			currentSize++
		}

		generateMaze()
		placeExit(1, 1)

		p.x = 1.5
		p.y = 1.5
		p.angle = 0
	}

	return nil
}

///////////////////////////////////////////////////////////////
// DRAW
///////////////////////////////////////////////////////////////

func (g *Game) Draw(screen *ebiten.Image) {

	// FINAL SCREEN
	if g.finished {

		screen.Fill(color.RGBA{15, 0, 25, 255})

		// More organic brain shape
		brain := color.RGBA{255, 120, 180, 255}
		shadow := color.RGBA{200, 80, 140, 255}

		ebitenutil.DrawRect(screen, 90, 50, 60, 70, brain)
		ebitenutil.DrawRect(screen, 150, 50, 60, 70, brain)
		ebitenutil.DrawRect(screen, 130, 80, 40, 30, brain)

		ebitenutil.DrawRect(screen, 80, 65, 40, 40, shadow)
		ebitenutil.DrawRect(screen, 170, 65, 40, 40, shadow)

		ebitenutil.DebugPrintAt(
			screen,
			"After all these years... you finally woke up, Terry",
			25, 150,
		)

		ebitenutil.DebugPrintAt(
			screen,
			"Press R to restart",
			100, 170,
		)

		return
	}

	width, height := screen.Size()
	fov := math.Pi / 3
	numRays := width

	for i := 0; i < numRays; i++ {

		rayAngle := g.player.angle - fov/2 +
			fov*float64(i)/float64(numRays)

		distance, cell := castRay(g.player.x, g.player.y, rayAngle)

		// FOG SYSTEM (visibility decreases each level)
		fogDistance := 20.0 - float64(g.level*3)
		if fogDistance < 6 {
			fogDistance = 6
		}

		if distance > fogDistance {
			continue
		}

		correctedDist := distance * math.Cos(rayAngle-g.player.angle)
		lineHeight := int(float64(height) / correctedDist)

		col := getWallColor(cell, distance, g.level)

		ebitenutil.DrawLine(
			screen,
			float64(i),
			float64(height/2-lineHeight/2),
			float64(i),
			float64(height/2+lineHeight/2),
			col,
		)
	}

	// Level info
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Level: %d", g.level), 10, 10)

	var message string

	switch g.level {
	case 1:
		message = "Escape probability extremely low (crying sounds)"
	case 2:
		message = "Brain death probability eliminated (sobbing)"
	case 3:
		message = "Cognitive functions improving (joy screams)"
	case 4:
		message = "..."
	}

	ebitenutil.DebugPrintAt(screen, message, 10, 30)
}

///////////////////////////////////////////////////////////////
// RAYCAST
///////////////////////////////////////////////////////////////

func castRay(px, py, angle float64) (float64, int) {

	step := 0.05
	distance := 0.0
	var cell int

	for {
		x := px + math.Cos(angle)*distance
		y := py + math.Sin(angle)*distance

		if int(x) < 0 || int(x) >= currentSize ||
			int(y) < 0 || int(y) >= currentSize {
			break
		}

		cell = gameMap[int(y)][int(x)]
		if cell == 1 || cell == 2 {
			break
		}

		distance += step
	}

	if distance == 0 {
		distance = 0.1
	}

	return distance, cell
}

///////////////////////////////////////////////////////////////
// WALL COLORS + FOG SHADING
///////////////////////////////////////////////////////////////

func getWallColor(cell int, distance float64, level int) color.RGBA {

	switch cell {

	case 1:
		shade := uint8(200 / (1 + distance*0.1))

		// Fog darkening
		fogFactor := 1.0 - (distance / 20.0)
		if fogFactor < 0 {
			fogFactor = 0
		}
		shade = uint8(float64(shade) * fogFactor)

		if level == 1 {
			return color.RGBA{shade - 10, shade, shade + 20, 255}
		} else if level == 2 {
			return color.RGBA{shade + 30, shade - 20, shade - 10, 255}
		} else if level == 3 {
			return color.RGBA{shade + 10, shade + 10, shade, 255}
		} else {
			// GOLD FINAL LEVEL
			return color.RGBA{
				uint8(float64(shade) * 1.2),
				uint8(float64(shade) * 1.0),
				uint8(float64(shade) * 0.3),
				255,
			}
		}

	case 2:
		// Final door blue
		if level == 4 {
			return color.RGBA{0, 120, 255, 255}
		}
		return color.RGBA{0, 255, 0, 255}

	default:
		return color.RGBA{0, 0, 0, 255}
	}
}

///////////////////////////////////////////////////////////////

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 200
}

///////////////////////////////////////////////////////////////

func main() {

	generateMaze()
	placeExit(1, 1)

	game := &Game{
		player: Player{x: 1.5, y: 1.5, angle: 0, speed: 0.1},
		level:  1,
	}

	ebiten.SetWindowSize(640, 400)
	ebiten.SetWindowTitle("IMPOSSIBLE BACKROOMS")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
