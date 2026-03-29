//go:build js && wasm

package main

import (
	"math"
	"math/rand"
	"syscall/js"
)

const (
	playerSpeed   = 0.15
	bulletSpeed   = 0.4
	enemySpeed    = 0.03
	spawnInterval = 60 // frames
)

type Vec3 struct {
	X, Y, Z float64
}

type Entity struct {
	Pos    Vec3
	Vel    Vec3
	Active bool
	HP     int
}

type Game struct {
	Player     Entity
	Bullets    []Entity
	Enemies    []Entity
	Score      int
	Frame      int
	Keys       map[string]bool
	ShootReady bool
}

var game *Game

func newGame() *Game {
	return &Game{
		Player: Entity{
			Pos:    Vec3{0, 0, 0},
			Active: true,
			HP:     3,
		},
		Bullets:    make([]Entity, 0, 100),
		Enemies:    make([]Entity, 0, 50),
		Keys:       make(map[string]bool),
		ShootReady: true,
	}
}

func (g *Game) Update() {
	g.Frame++

	// Player movement
	dx, dz := 0.0, 0.0
	if g.Keys["KeyW"] || g.Keys["ArrowUp"] {
		dz = -playerSpeed
	}
	if g.Keys["KeyS"] || g.Keys["ArrowDown"] {
		dz = playerSpeed
	}
	if g.Keys["KeyA"] || g.Keys["ArrowLeft"] {
		dx = -playerSpeed
	}
	if g.Keys["KeyD"] || g.Keys["ArrowRight"] {
		dx = playerSpeed
	}
	g.Player.Pos.X += dx
	g.Player.Pos.Z += dz

	// Clamp player position
	g.Player.Pos.X = clamp(g.Player.Pos.X, -8, 8)
	g.Player.Pos.Z = clamp(g.Player.Pos.Z, -8, 8)

	// Shoot
	if g.Keys["Space"] && g.ShootReady {
		g.Bullets = append(g.Bullets, Entity{
			Pos:    Vec3{g.Player.Pos.X, 0.5, g.Player.Pos.Z},
			Vel:    Vec3{0, 0, -bulletSpeed},
			Active: true,
		})
		g.ShootReady = false
	}
	if !g.Keys["Space"] {
		g.ShootReady = true
	}

	// Update bullets
	for i := range g.Bullets {
		if !g.Bullets[i].Active {
			continue
		}
		g.Bullets[i].Pos.X += g.Bullets[i].Vel.X
		g.Bullets[i].Pos.Z += g.Bullets[i].Vel.Z
		if g.Bullets[i].Pos.Z < -15 {
			g.Bullets[i].Active = false
		}
	}

	// Spawn enemies
	if g.Frame%spawnInterval == 0 {
		g.Enemies = append(g.Enemies, Entity{
			Pos:    Vec3{rand.Float64()*16 - 8, 0, -12},
			Vel:    Vec3{0, 0, enemySpeed},
			Active: true,
			HP:     1,
		})
	}

	// Update enemies
	for i := range g.Enemies {
		if !g.Enemies[i].Active {
			continue
		}
		g.Enemies[i].Pos.Z += g.Enemies[i].Vel.Z
		if g.Enemies[i].Pos.Z > 10 {
			g.Enemies[i].Active = false
		}
	}

	// Collision: bullets vs enemies
	for i := range g.Bullets {
		if !g.Bullets[i].Active {
			continue
		}
		for j := range g.Enemies {
			if !g.Enemies[j].Active {
				continue
			}
			if dist(g.Bullets[i].Pos, g.Enemies[j].Pos) < 1.0 {
				g.Bullets[i].Active = false
				g.Enemies[j].Active = false
				g.Score += 100
			}
		}
	}

	// Collision: enemies vs player
	for i := range g.Enemies {
		if !g.Enemies[i].Active {
			continue
		}
		if dist(g.Enemies[i].Pos, g.Player.Pos) < 1.2 {
			g.Enemies[i].Active = false
			g.Player.HP--
		}
	}
}

func (g *Game) ToJS() js.Value {
	obj := js.Global().Get("Object").New()
	obj.Set("playerX", g.Player.Pos.X)
	obj.Set("playerZ", g.Player.Pos.Z)
	obj.Set("playerHP", g.Player.HP)
	obj.Set("score", g.Score)

	bullets := js.Global().Get("Array").New()
	for _, b := range g.Bullets {
		if !b.Active {
			continue
		}
		bo := js.Global().Get("Object").New()
		bo.Set("x", b.Pos.X)
		bo.Set("z", b.Pos.Z)
		bullets.Call("push", bo)
	}
	obj.Set("bullets", bullets)

	enemies := js.Global().Get("Array").New()
	for _, e := range g.Enemies {
		if !e.Active {
			continue
		}
		eo := js.Global().Get("Object").New()
		eo.Set("x", e.Pos.X)
		eo.Set("z", e.Pos.Z)
		enemies.Call("push", eo)
	}
	obj.Set("enemies", enemies)

	return obj
}

func dist(a, b Vec3) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dz*dz)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func main() {
	game = newGame()

	js.Global().Set("goGameUpdate", js.FuncOf(func(this js.Value, args []js.Value) any {
		game.Update()
		return game.ToJS()
	}))

	js.Global().Set("goKeyDown", js.FuncOf(func(this js.Value, args []js.Value) any {
		game.Keys[args[0].String()] = true
		return nil
	}))

	js.Global().Set("goKeyUp", js.FuncOf(func(this js.Value, args []js.Value) any {
		game.Keys[args[0].String()] = false
		return nil
	}))

	js.Global().Set("goResetGame", js.FuncOf(func(this js.Value, args []js.Value) any {
		game = newGame()
		return nil
	}))

	// Keep alive
	select {}
}
