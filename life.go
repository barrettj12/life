package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	WIDTH  = 32
	HEIGHT = 13
	SLEEP  = 200 * time.Millisecond
)

func main() {
	// get delay from args
	// if not smooth, increase the delay
	if len(os.Args) > 1 {
		delay, err := time.ParseDuration(os.Args[1])
		if err == nil {
			SLEEP = delay
		}
	}

	setWH()
	board := make([][]bool, HEIGHT)
	for y := range board {
		board[y] = make([]bool, WIDTH)
	}

	// Initial state of board
	board[4][5] = true
	board[5][5] = true
	board[5][6] = true
	board[5][7] = true
	board[6][6] = true

	boardsCh := make(chan string, 10)
	bp := boardPrinter{boardsCh}
	go bp.listen()

	for {
		boardsCh <- render(board)
		tick(board)
	}
}

func render(board [][]bool) string {
	// Render screen first, print all at once
	screen := "\033[H\033[2J"
	for i, row := range board {
		for _, sq := range row {
			if sq {
				screen += "  "
			} else {
				screen += "██"
			}
		}
		if i+1 < len(board) {
			screen += "\n"
		}
	}

	return screen
}

func tick(board [][]bool) {
	// Calculate live neighbours first
	liveNeighbours := make([][]int, HEIGHT)
	for y := range liveNeighbours {
		liveNeighbours[y] = make([]int, WIDTH)
	}

	for y := range board {
		for x := range board[y] {
			if board[y][x] {
				for _, d := range NEIGHBOUR_SHIFTS {
					nx := getX(x + d.x)
					ny := getY(y + d.y)
					liveNeighbours[ny][nx]++
				}
			}
		}
	}

	// Update board
	for y := range board {
		for x := range board[y] {
			if board[y][x] && liveNeighbours[y][x] < 2 {
				board[y][x] = false
			}
			if board[y][x] && liveNeighbours[y][x] > 3 {
				board[y][x] = false
			}
			if !board[y][x] && liveNeighbours[y][x] == 3 {
				board[y][x] = true
			}
		}
	}
}

var NEIGHBOUR_SHIFTS = []point{
	{-1, -1}, {-1, 0}, {-1, 1}, {0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1},
}

type point struct{ x, y int }

func getX(x int) int {
	if x < 0 {
		return x + WIDTH
	} else if x >= WIDTH {
		return x - WIDTH
	}
	return x
}

func getY(y int) int {
	if y < 0 {
		return y + HEIGHT
	} else if y >= HEIGHT {
		return y - HEIGHT
	}
	return y
}

func setWH() {
	bw, _ := exec.Command("tput", "cols").Output()
	w, _ := strconv.Atoi(strings.TrimSpace(string(bw)))
	WIDTH = w / 2
	bh, _ := exec.Command("tput", "lines").Output()
	HEIGHT, _ = strconv.Atoi(strings.TrimSpace(string(bh)))
}

// Board printer
// In separate goroutine so that ticking is smooth
type boardPrinter struct {
	boards chan string
}

func (b *boardPrinter) listen() {
	ticker := time.NewTicker(SLEEP)
	for {
		<-ticker.C
		fmt.Print(<-b.boards)
	}
}
