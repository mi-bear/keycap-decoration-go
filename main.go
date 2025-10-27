//go:build wioterminal

package main

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/sago35/koebiten"
	"github.com/sago35/koebiten/hardware"
	"tinygo.org/x/drivers/pixel"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/shnm"
)

const (
	logicalW = 128
	logicalH = 64
)

type AppState uint8

const (
	StateMood AppState = iota
	StatePart
	StateResult
)

type Mood uint8

const (
	MoodHigh Mood = iota
	MoodMid
	MoodLow
)

type PartGroup uint8

const (
	PartCute PartGroup = iota
	PartSimple
	PartCool
)

var (
	white = pixel.NewMonochrome(0xFF, 0xFF, 0xFF)
	black = pixel.NewMonochrome(0x00, 0x00, 0x00)
)

var (
	// 今日の気分 = ベースカラー番号
	moodToBaseNums = map[Mood][]int{
		MoodHigh: makeRange(1, 8),
		MoodMid:  makeRange(9, 16),
		MoodLow:  makeRange(17, 24),
	}

	// キーキャップのパーツ番号
	partToBaseNums = map[PartGroup][]int{
		PartCute:   makeRange(1, 3),
		PartSimple: makeRange(4, 6),
		PartCool:   makeRange(7, 9),
	}
)

type Game struct {
	state     AppState
	mood      Mood
	partGroup PartGroup
	baseNum   int
	partNum   int

	frame   int
	rocketX int
}

func (g *Game) Update() error {
	g.frame++
	switch g.state {
	case StateMood:
		if koebiten.IsKeyJustPressed(koebiten.Key0) {
			g.mood = MoodHigh
			g.baseNum = pickNum(moodToBaseNums[g.mood])
			g.state = StatePart
			g.rocketX = -20
		}
		if koebiten.IsKeyJustPressed(koebiten.Key1) {
			g.mood = MoodMid
			g.baseNum = pickNum(moodToBaseNums[g.mood])
			g.state = StatePart
			g.rocketX = -20
		}
		if koebiten.IsKeyJustPressed(koebiten.Key2) {
			g.mood = MoodLow
			g.baseNum = pickNum(moodToBaseNums[g.mood])
			g.state = StatePart
			g.rocketX = -20
		}
	case StatePart:
		if koebiten.IsKeyJustPressed(koebiten.Key0) {
			g.partGroup = PartCute
			g.partNum = pickNum(partToBaseNums[g.partGroup])
			g.state = StateResult
		}
		if koebiten.IsKeyJustPressed(koebiten.Key1) {
			g.partGroup = PartSimple
			g.partNum = pickNum(partToBaseNums[g.partGroup])
			g.state = StateResult
		}
		if koebiten.IsKeyJustPressed(koebiten.Key2) {
			g.partGroup = PartCool
			g.partNum = pickNum(partToBaseNums[g.partGroup])
			g.state = StateResult
		}

		g.rocketX += 2
		if g.rocketX > logicalW+18 {
			g.rocketX = -18
		}
	case StateResult:
		if koebiten.IsKeyJustPressed(koebiten.KeyDown) {
			*g = Game{state: StateMood}
		}
	}
	return nil
}

func (g *Game) Draw(screen *koebiten.Image) {
	koebiten.DrawRect(screen, 0, 0, logicalW, logicalH, white)

	switch g.state {
	case StateMood:
		koebiten.DrawText(screen, "今日の気分は？", &shnm.Shnmk12, 2, 15, white)
		top := 26
		drawThreeButtons(screen, []string{"High", "Mid", "Low"}, top)
		drawMoodStars(screen, top)

	case StatePart:
		koebiten.DrawText(screen, "パーツのイメージは？", &shnm.Shnmk12, 2, 15, white)
		top := 26
		drawThreeButtons(screen, []string{"Cute", "Simple", "Cool"}, top)
		yCenter := top + 30
		drawRocket(screen, g.rocketX, yCenter, g.frame)

	case StateResult:
		koebiten.DrawText(screen, "キミにきめた！", &shnm.Shnmk12, 2, 15, white)

		// [mood] & [part]
		expr := "[" + strconv.Itoa(g.baseNum) + "] & [" + strconv.Itoa(g.partNum) + "]"
		koebiten.DrawText(screen, expr, &shnm.Shnmk12, 2, 40, white)

		// 踊るクマ
		bearCX, bearCY := 104, 40
		drawKumaDancing(screen, bearCX, bearCY, g.frame)

		// リスタート
		koebiten.DrawText(screen, "[Down] Restart", &tinyfont.Org01, 2, logicalH-5, white)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return logicalW, logicalH
}

func init() {
	v := machine.ADC{Pin: machine.A0}
	v.Configure()

	seed := time.Now().UnixNano() ^
		int64(v.Get()) ^
		int64(time.Now().Unix())

	rand.Seed(seed)
}

func main() {
	koebiten.SetHardware(hardware.Device)
	if err := koebiten.RunGame(&Game{state: StateMood}); err != nil {
		panic(err)
	}
}

func makeRange(a, b int) []int {
	if b < a {
		return nil
	}
	out := make([]int, 0, b-a+1)
	for i := a; i <= b; i++ {
		out = append(out, i)
	}
	return out
}

func pickNum(v []int) int { return v[rand.Intn(len(v))] }

func drawThreeButtons(screen *koebiten.Image, labels []string, top int) {
	w, h := 38, 18
	left := 4
	gap := 2
	for i := 0; i < 3; i++ {
		x := left + i*(w+gap)
		koebiten.DrawRect(screen, x, top, w, h, white)
		koebiten.DrawText(screen, labels[i], &tinyfont.Org01, int16(x+6), int16(top+h/2+2), white)
	}
}

func drawMoodStars(screen *koebiten.Image, top int) {
	w, h := 38, 18
	left := 4
	gap := 2
	iconY := top + h + 4
	r := 3
	for i := 0; i < 3; i++ {
		x := left + i*(w+gap)
		cx := x + w/2
		count := 3 - i // 左 3, 中央 2, 右 1
		for s := 0; s < count; s++ {
			drawStarIcon(screen, cx-(count-1)*4+s*8, iconY, r)
		}
	}
}

func drawStarIcon(screen *koebiten.Image, cx, cy, r int) {
	if r < 3 {
		return
	}
	pts := make([][2]int, 5)
	for i := 0; i < 5; i++ {
		ang := (-90.0 + float64(i)*72.0) * math.Pi / 180.0
		x := cx + int(float64(r)*math.Cos(ang))
		y := cy + int(float64(r)*math.Sin(ang))
		pts[i] = [2]int{x, y}
	}
	for i := 0; i < 5; i++ {
		j := (i + 2) % 5
		koebiten.DrawLine(screen, pts[i][0], pts[i][1], pts[j][0], pts[j][1], white)
	}
}

func drawRocket(screen *koebiten.Image, x, y, frame int) {
	bodyW, bodyH := 20, 10
	left := x - bodyW/2
	right := x + bodyW/2
	top := y - bodyH/2
	bot := y + bodyH/2

	koebiten.DrawRect(screen, left, top, bodyW, bodyH, white)
	koebiten.DrawTriangle(screen, right, y, right-6, top, right-6, bot, white)
	koebiten.DrawTriangle(screen, left, y, left+5, top+2, left+5, bot-2, white)

	winCX, winCY := x-2, y
	drawCircleIcon(screen, winCX, winCY, 2)

	flameLen := 6
	if (frame/4)%2 == 1 {
		flameLen = 8
	}
	koebiten.DrawLine(screen, left-1, y, left-flameLen, y-3, white)
	koebiten.DrawLine(screen, left-1, y, left-flameLen, y+3, white)
}

func drawCircleIcon(screen *koebiten.Image, cx, cy, r int) {
	const step = 8
	for deg := 0; deg < 360; deg += step {
		rad := float64(deg) * math.Pi / 180
		x := cx + int(float64(r)*math.Cos(rad))
		y := cy + int(float64(r)*math.Sin(rad))
		koebiten.DrawRect(screen, x, y, 1, 1, white)
	}
}

// 踊るクマ
func drawKumaDancing(screen *koebiten.Image, cx, cy, frame int) {
	headR := 7
	earR := 3
	bodyW := 12
	bodyH := 10

	if (frame/6)%2 == 1 {
		cy++
	}

	// クマの頭 & 耳
	koebiten.DrawCircle(screen, cx, cy-8, headR, white)
	koebiten.DrawCircle(screen, cx-7, cy-14, earR, white)
	koebiten.DrawCircle(screen, cx+7, cy-14, earR, white)

	// クマの顔
	koebiten.DrawRect(screen, cx-2, cy-9, 1, 1, white)
	koebiten.DrawRect(screen, cx+2, cy-9, 1, 1, white)
	koebiten.DrawRect(screen, cx, cy-7, 1, 1, white)
	koebiten.DrawLine(screen, cx-1, cy-6, cx+1, cy-6, white)

	// クマの身体
	drawRoundedBody(screen, cx, cy)

	// クマの手足 (アニメーション風)
	wave := (frame/8)%2 == 0

	if wave {
		koebiten.DrawLine(screen, cx+bodyW/2, cy, cx+bodyW/2+4, cy-4, white)
		koebiten.DrawLine(screen, cx-bodyW/2, cy+3, cx-bodyW/2-4, cy+5, white)
	} else {
		koebiten.DrawLine(screen, cx+bodyW/2, cy+3, cx+bodyW/2+4, cy+5, white)
		koebiten.DrawLine(screen, cx-bodyW/2, cy, cx-bodyW/2-4, cy-4, white)
	}

	if wave {
		koebiten.DrawLine(screen, cx-3, cy+bodyH, cx-5, cy+bodyH+3, white)
		koebiten.DrawLine(screen, cx+3, cy+bodyH, cx+6, cy+bodyH+2, white)
	} else {
		koebiten.DrawLine(screen, cx-3, cy+bodyH, cx-6, cy+bodyH+2, white)
		koebiten.DrawLine(screen, cx+3, cy+bodyH, cx+5, cy+bodyH+3, white)
	}
}

// クマの身体
func drawRoundedBody(screen *koebiten.Image, cx, cy int) {
	bodyW := 12
	bodyH := 10
	r := 2 // 丸み半径

	left := cx - bodyW/2
	top := cy - 2
	right := left + bodyW
	bottom := top + bodyH

	// 横線
	koebiten.DrawLine(screen, left+r, top, right-r, top, white)
	koebiten.DrawLine(screen, left+r, bottom, right-r, bottom, white)

	// 縦線
	koebiten.DrawLine(screen, left, top+r, left, bottom-r, white)
	koebiten.DrawLine(screen, right, top+r, right, bottom-r, white)

	// 角の丸み
	koebiten.DrawRect(screen, left+1, top+1, 1, 1, white)
	koebiten.DrawRect(screen, right-1, top+1, 1, 1, white)
	koebiten.DrawRect(screen, left+1, bottom-1, 1, 1, white)
	koebiten.DrawRect(screen, right-1, bottom-1, 1, 1, white)
}
