package casino

import (
	"math"
)

const (
	SweetReels  = 5
	SweetRows   = 5
	SweetCells  = SweetReels * SweetRows
)

// sweetEngine — изолированный движок Sweet Bonanza
type sweetEngine struct {
	fairness *Fairness
	idCounter int64
}

// Веса символов (чем выше — тем чаще выпадает)
var sweetWeights = []struct {
	symbol string
	weight int
}{
	{"strawberry", 12},
	{"pear", 11},
	{"orange", 10},
	{"blueberry", 9},
	{"apple", 8},
	{"watermelon", 7},
	{"grape", 6},
	{"heart", 5},
	{"scatter", 3}, // lollipop
}

// Paytable: минимальное кол-во символов -> множитель от ставки.
// RTP ~71% base (без фриспинов), с фриспинами ~95-96%.
// Пороги: редкие символы (heart, grape) платят с 6+, остальные с 7+.
var sweetPays = map[string]map[int]int64{
	"heart":      {6: 4, 7: 6, 8: 12, 9: 20, 10: 35, 11: 55, 12: 90},
	"grape":      {6: 2, 7: 5, 8: 8, 9: 15, 10: 25, 11: 40, 12: 70},
	"watermelon": {7: 1, 8: 3, 9: 6, 10: 12, 11: 20, 12: 35},
	"apple":      {7: 1, 8: 2, 9: 5, 10: 8, 11: 15, 12: 25},
	"blueberry":  {7: 1, 8: 1, 9: 3, 10: 6, 11: 12, 12: 20},
	"orange":     {7: 1, 8: 1, 9: 2, 10: 5, 11: 8, 12: 16},
	"pear":       {7: 1, 8: 1, 9: 1, 10: 4, 11: 6, 12: 12},
	"strawberry": {7: 1, 8: 1, 9: 1, 10: 3, 11: 5, 12: 10},
}

// Каскадные множители (1-based cascade index).
// Первый каскад x1, второй x1, третий x1, четвёртый x2, пятый x3, далее x3.
var cascadeMultipliers = []int64{1, 1, 1, 2, 3}

// Bomb values and weights (free spins only).
// Снижены для целевого RTP ~95-96% (фриспины ~24%, база ~71%).
var bombValues = []int64{2, 3, 4, 5, 6, 8, 10, 12, 15, 20}
var bombWeights = []int{30, 25, 20, 15, 10, 8, 6, 4, 3, 2}

const (
	maxCascades      = 50 // hard cap
	freeSpinsTrigger = 4  // scatters needed
	freeSpinsAward   = 10
	freeSpinsRetrigger = 3
	freeSpinsRetriggerAward = 5
	bombSpawnChance = 0.02 // 2% per cell in free spins
)

func newSweetEngine(serverSeed, clientSeed string, nonce int64) *sweetEngine {
	return &sweetEngine{
		fairness: NewFairness(serverSeed, clientSeed, nonce),
		idCounter: 1,
	}
}

func (e *sweetEngine) draw(max int) int {
	return DrawIntWithFairness(e.fairness, max)
}

func (e *sweetEngine) nextID() int64 {
	id := e.idCounter
	e.idCounter++
	return id
}

// symbol — выбор символа для ячейки
func (e *sweetEngine) symbol(freeSpins bool) SweetCell {
	// Bomb spawn in free spins
	if freeSpins && e.draw(1000) < int(bombSpawnChance*1000) {
		return SweetCell{
			ID:     e.nextID(),
			Symbol: "bomb",
			Mult:   e.bombValue(),
		}
	}
	total := 0
	for _, w := range sweetWeights {
		total += w.weight
	}
	r := e.draw(total)
	for _, w := range sweetWeights {
		r -= w.weight
		if r < 0 {
			return SweetCell{ID: e.nextID(), Symbol: w.symbol}
		}
	}
	return SweetCell{ID: e.nextID(), Symbol: "strawberry"}
}

func (e *sweetEngine) bombValue() int64 {
	total := 0
	for _, w := range bombWeights {
		total += w
	}
	r := e.draw(total)
	for i, w := range bombWeights {
		r -= w
		if r <= 0 {
			return bombValues[i]
		}
	}
	return bombValues[0]
}

// board — генерация 5×5 доски
func (e *sweetEngine) board(freeSpins bool) [][]SweetCell {
	board := make([][]SweetCell, SweetReels)
	for r := 0; r < SweetReels; r++ {
		board[r] = make([]SweetCell, SweetRows)
		for row := 0; row < SweetRows; row++ {
			board[r][row] = e.symbol(freeSpins)
		}
	}
	return board
}

func copySweetBoard(board [][]SweetCell) [][]SweetCell {
	copyBoard := make([][]SweetCell, len(board))
	for r := range board {
		copyBoard[r] = make([]SweetCell, len(board[r]))
		copy(copyBoard[r], board[r])
	}
	return copyBoard
}

// evaluation — pay-anywhere подсчёт выигрыша
func (e *sweetEngine) evaluate(board [][]SweetCell, bet int64) ([]SweetPosition, int64) {
	counts := make(map[string]int)
	positions := make(map[string][]SweetPosition)

	for r := 0; r < SweetReels; r++ {
		for row := 0; row < SweetRows; row++ {
			cell := board[r][row]
			if cell.Symbol == "bomb" || cell.Symbol == "scatter" {
				continue
			}
			counts[cell.Symbol]++
			positions[cell.Symbol] = append(positions[cell.Symbol], SweetPosition{Reel: r, Row: row})
		}
	}

	winning := make([]SweetPosition, 0)
	var multiplier int64

	for symbol, count := range counts {
		pay := sweetPays[symbol]
		if count < 5 || pay == nil {
			continue
		}
		best := int64(0)
		for threshold, mult := range pay {
			if count >= threshold && mult > best {
				best = mult
			}
		}
		if best > 0 {
			winning = append(winning, positions[symbol]...)
			multiplier += best
		}
	}

	return winning, multiplier * bet
}

func cascadeMult(cascade int) int64 {
	if cascade < len(cascadeMultipliers) {
		return cascadeMultipliers[cascade]
	}
	return cascadeMultipliers[len(cascadeMultipliers)-1]
}

// tumble — гравитация: выигрышные удаляются, остальные падают вниз, новые сверху
func (e *sweetEngine) tumble(board [][]SweetCell, winning []SweetPosition, freeSpins bool) [][]SweetCell {
	remove := make(map[[2]int]bool, len(winning))
	for _, pos := range winning {
		remove[[2]int{pos.Reel, pos.Row}] = true
	}

	next := make([][]SweetCell, SweetReels)
	for r := 0; r < SweetReels; r++ {
		kept := make([]SweetCell, 0, SweetRows)
		for row := 0; row < SweetRows; row++ {
			if !remove[[2]int{r, row}] {
				kept = append(kept, board[r][row])
			}
		}
		col := make([]SweetCell, 0, SweetRows)
		for len(col)+len(kept) < SweetRows {
			col = append(col, e.symbol(freeSpins))
		}
		next[r] = append(col, kept...)
	}
	return next
}

func countScatters(board [][]SweetCell) int {
	n := 0
	for _, col := range board {
		for _, cell := range col {
			if cell.Symbol == "scatter" {
				n++
			}
		}
	}
	return n
}

func bombMult(board [][]SweetCell) int64 {
	var total int64
	for _, col := range board {
		for _, cell := range col {
			if cell.Symbol == "bomb" {
				total += cell.Mult
			}
		}
	}
	return total
}

// Spin — основной метод: генерирует полный timeline спина
func (e *sweetEngine) Spin(bet int64, freeSpins bool) SweetBonanzaResult {
	board := e.board(freeSpins)
	initial := copySweetBoard(board)
	result := SweetBonanzaResult{
		InitialBoard: initial,
		ScatterCount: countScatters(board),
	}
	var totalWin int64

	for cascade := 0; cascade < maxCascades; cascade++ {
		winning, rawWin := e.evaluate(board, bet)
		if rawWin == 0 {
			break
		}
		mult := cascadeMult(cascade + 1)
		stepWin := rawWin * mult
		totalWin += stepWin

		result.Steps = append(result.Steps, SweetCascadeStep{
			Board:            copySweetBoard(board),
			WinningPositions: winning,
			Cascade:          cascade + 1,
			Multiplier:       mult,
			Win:              stepWin,
		})

		board = e.tumble(board, winning, freeSpins)
	}

	result.FinalBoard = copySweetBoard(board)

	if freeSpins && totalWin > 0 {
		if bomb := bombMult(board); bomb > 0 {
			result.BombMultiplier = bomb
			totalWin *= bomb
		}
	}

	if result.ScatterCount >= freeSpinsTrigger {
		if freeSpins {
			result.FreeSpinsAwarded = freeSpinsRetriggerAward
		} else {
			result.FreeSpinsAwarded = freeSpinsAward
		}
	}

	result.TotalWin = int64(math.Max(0, float64(totalWin)))
	return result
}