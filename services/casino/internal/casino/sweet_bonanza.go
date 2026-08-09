package casino

import "math"

const (
	sweetReels = 5
	sweetRows  = 5
)

type SweetCell struct {
	ID     int64  `json:"id"`
	Symbol string `json:"symbol"`
	Mult   int64  `json:"mult,omitempty"`
}

type SweetPosition struct {
	Reel int `json:"reel"`
	Row  int `json:"row"`
}

type SweetCascadeStep struct {
	Board            [][]SweetCell   `json:"board"`
	WinningPositions []SweetPosition `json:"winning_positions"`
	Cascade          int             `json:"cascade"`
	Multiplier       int64           `json:"multiplier"`
	Win              int64           `json:"win"`
}

type SweetBonanzaResult struct {
	InitialBoard     [][]SweetCell      `json:"initial_board"`
	Steps            []SweetCascadeStep `json:"steps"`
	FinalBoard       [][]SweetCell      `json:"final_board"`
	ScatterCount     int                `json:"scatter_count"`
	BombMultiplier   int64              `json:"bomb_multiplier,omitempty"`
	FreeSpinsAwarded int64              `json:"free_spins_awarded,omitempty"`
	TotalWin         int64              `json:"total_win"`
}

type sweetEngine struct {
	fairness *Fairness
	nextID   int64
}

var sweetWeights = []struct {
	symbol string
	weight int
}{
	{"strawberry", 12}, {"pear", 11}, {"orange", 10}, {"blueberry", 9},
	{"apple", 8}, {"watermelon", 7}, {"grape", 6}, {"heart", 5}, {"scatter", 3},
}

const sweetMaxPayoutMultiplier int64 = 5000

var sweetPays = map[string]map[int]int64{
	// The old frontend table was visual-only and had no RTP controls. These
	// server values pay from 9+ symbols and are calibrated for real credits.
	"heart":      {9: 48, 10: 90, 11: 150, 12: 300},
	"grape":      {9: 36, 10: 72, 11: 120, 12: 210},
	"watermelon": {9: 30, 10: 54, 11: 90, 12: 168},
	"apple":      {9: 24, 10: 42, 11: 72, 12: 132},
	"blueberry":  {9: 18, 10: 36, 11: 60, 12: 108},
	"orange":     {9: 18, 10: 30, 11: 48, 12: 90},
	"pear":       {9: 12, 10: 24, 11: 42, 12: 72},
	"strawberry": {9: 12, 10: 18, 11: 30, 12: 60},
}

func newSweetEngine(serverSeed, clientSeed string, nonce int64) *sweetEngine {
	return &sweetEngine{fairness: NewFairness(serverSeed, clientSeed, nonce)}
}

func (e *sweetEngine) draw(max int) int { return DrawIntWithFairness(e.fairness, max) }

func (e *sweetEngine) symbol(freeSpins bool) SweetCell {
	e.nextID++
	id := e.nextID
	if freeSpins && e.draw(100) < 8 {
		bombs := []int64{2, 3, 4, 5, 6, 8, 10, 15, 20, 25, 50}
		weights := []int{26, 22, 18, 14, 10, 8, 6, 4, 3, 2, 1}
		total := 0
		for _, weight := range weights {
			total += weight
		}
		r := e.draw(total)
		for i, weight := range weights {
			r -= weight
			if r < 0 {
				return SweetCell{ID: id, Symbol: "bomb", Mult: bombs[i]}
			}
		}
	}
	total := 0
	for _, item := range sweetWeights {
		total += item.weight
	}
	r := e.draw(total)
	for _, item := range sweetWeights {
		r -= item.weight
		if r < 0 {
			return SweetCell{ID: id, Symbol: item.symbol}
		}
	}
	return SweetCell{ID: id, Symbol: "strawberry"}
}

func (e *sweetEngine) board(freeSpins bool) [][]SweetCell {
	board := make([][]SweetCell, sweetReels)
	for reel := range board {
		board[reel] = make([]SweetCell, sweetRows)
		for row := range board[reel] {
			board[reel][row] = e.symbol(freeSpins)
		}
	}
	return board
}

func copySweetBoard(board [][]SweetCell) [][]SweetCell {
	copyBoard := make([][]SweetCell, len(board))
	for reel := range board {
		copyBoard[reel] = append([]SweetCell(nil), board[reel]...)
	}
	return copyBoard
}

func sweetEvaluation(board [][]SweetCell, bet int64) ([]SweetPosition, int64) {
	positions := make(map[string][]SweetPosition)
	for reel := range board {
		for row, cell := range board[reel] {
			if cell.Symbol != "scatter" && cell.Symbol != "bomb" {
				positions[cell.Symbol] = append(positions[cell.Symbol], SweetPosition{Reel: reel, Row: row})
			}
		}
	}
	winning := make([]SweetPosition, 0)
	var multiplier int64
	for symbol, entries := range positions {
		pay := sweetPays[symbol]
		if len(entries) < 5 || pay == nil {
			continue
		}
		best := int64(0)
		for count, value := range pay {
			if len(entries) >= count && value > best {
				best = value
			}
		}
		if best > 0 {
			winning = append(winning, entries...)
			multiplier += best
		}
	}
	return winning, multiplier * bet
}

func sweetCascadeMultiplier(cascade int) int64 {
	values := []int64{1, 1, 2, 3, 5, 8}
	if cascade < len(values) {
		return values[cascade]
	}
	return values[len(values)-1]
}

func (e *sweetEngine) tumble(board [][]SweetCell, winning []SweetPosition, freeSpins bool) [][]SweetCell {
	remove := make(map[[2]int]bool, len(winning))
	for _, position := range winning {
		remove[[2]int{position.Reel, position.Row}] = true
	}
	next := make([][]SweetCell, sweetReels)
	for reel := range board {
		kept := make([]SweetCell, 0, sweetRows)
		for row, cell := range board[reel] {
			if !remove[[2]int{reel, row}] {
				kept = append(kept, cell)
			}
		}
		column := make([]SweetCell, 0, sweetRows)
		for len(column)+len(kept) < sweetRows {
			column = append(column, e.symbol(freeSpins))
		}
		next[reel] = append(column, kept...)
	}
	return next
}

func sweetScatters(board [][]SweetCell) int {
	count := 0
	for _, column := range board {
		for _, cell := range column {
			if cell.Symbol == "scatter" {
				count++
			}
		}
	}
	return count
}

func sweetBombMultiplier(board [][]SweetCell) int64 {
	var total int64
	for _, column := range board {
		for _, cell := range column {
			if cell.Symbol == "bomb" {
				total += cell.Mult
			}
		}
	}
	return total
}

func (e *sweetEngine) Spin(bet int64, freeSpins bool) SweetBonanzaResult {
	board := e.board(freeSpins)
	initial := copySweetBoard(board)
	result := SweetBonanzaResult{InitialBoard: initial, ScatterCount: sweetScatters(board)}
	var totalWin int64
	// A hard cap protects settlement from pathological all-winning boards.
	for cascade := 0; cascade < 50; cascade++ {
		winning, rawWin := sweetEvaluation(board, bet)
		if rawWin == 0 {
			break
		}
		multiplier := sweetCascadeMultiplier(cascade)
		stepWin := rawWin * multiplier
		totalWin += stepWin
		result.Steps = append(result.Steps, SweetCascadeStep{
			Board: copySweetBoard(board), WinningPositions: winning, Cascade: cascade + 1, Multiplier: multiplier, Win: stepWin,
		})
		board = e.tumble(board, winning, freeSpins)
	}
	result.FinalBoard = copySweetBoard(board)
	if freeSpins && totalWin > 0 {
		if bomb := sweetBombMultiplier(board); bomb > 0 {
			result.BombMultiplier = bomb
			totalWin *= bomb
		}
	}
	if result.ScatterCount >= 4 {
		if freeSpins {
			result.FreeSpinsAwarded = 5
		} else {
			result.FreeSpinsAwarded = 10
		}
	}
	if totalWin > sweetMaxPayoutMultiplier*bet {
		totalWin = sweetMaxPayoutMultiplier * bet
	}
	result.TotalWin = int64(math.Max(0, float64(totalWin)))
	return result
}
