package snake

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/gdamore/tcell"
)

const (
	Up = iota
	Left
	Right
	Down
)

type Food struct {
	Coordinates Coordinate
	Letter      string
	Point       int
	//TODO points, type, etc
}

type Game struct {
	mu             sync.Mutex
	Screen         tcell.Screen
	IsStart        bool
	IsOver         bool
	IsPaused       bool
	Food           []Food
	Board          *Board
	Speed          time.Duration
	Snakes         []*Snake
	PlayerNumber   int
	FoodNumber     int
	BotNumber      int
	whoLost        int
	BotPaths       map[int][]Coordinate
	settings       PlayersControlSettings
	TestFieldError string
	TestFields     []string
}

func StartGame(playerNumber int, foodNumber int, botNumber int) {
	playerDirChan := make([]chan int, 0)
	for i := 0; i < playerNumber; i++ {
		playerDirChan = append(playerDirChan, make(chan int, 1))
	}

	botDirChan := make([]chan int, 0)
	botRunBotChan := make([]chan bool, 0)
	for i := 0; i < botNumber; i++ {
		botDirChan = append(botDirChan, make(chan int, 1))
		botRunBotChan = append(botRunBotChan, make(chan bool, 1))
	}

	game := newGame(newBoard(50, 20), playerNumber, foodNumber, botNumber)

	// go game.Run(directionChan1, directionChan2, directionChanBot1, runBotCalcChan1, directionChanBot2, runBotCalcChan2)
	// go game.Run(directionChan1, directionChan2, botDirChan[0], botRunBotChan[0], directionChanBot2, runBotCalcChan2)
	go game.Run2(playerDirChan, botDirChan, botRunBotChan)
	go game.handleKeyBoardEvents(playerDirChan)

	// for i := playerNumber; i < len(game.Snakes); i++ {
	// 	go game.botControl(game.Snakes[i], botDirChan[i-playerNumber], botRunBotChan[i-playerNumber])
	// }

	//összes botnak egy chanel?
	//go game.botControl(game.Snakes[0], directionChanBot1, runBotCalcChan1)
	//go game.botControl(game.Snakes[1], directionChanBot2, runBotCalcChan2)

	for i := 0; i < botNumber; i++ {
		go game.botControl(game.Snakes[i+playerNumber], botDirChan[i], botRunBotChan[i], i+playerNumber)
	}

	// go game.botControl(game.Snakes[0], botDirChan[0], botRunBotChan[0])

	for {
	}
}

func newGame(board *Board, playerNumber int, foodNumber int, botNumber int) *Game {
	screen, err := tcell.NewScreen()

	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)

	game := &Game{
		Board:        board,
		Screen:       screen,
		Speed:        500 * time.Millisecond,
		Snakes:       make([]*Snake, 0),
		Food:         make([]Food, 0),
		PlayerNumber: playerNumber,
		FoodNumber:   foodNumber,
		BotNumber:    botNumber,
		TestFields:   make([]string, 0),
		BotPaths:     make(map[int][]Coordinate),
	}
	game.settings = controlls("playerControlSettings.json")

	game.TestFieldError = ""
	for i := 0; i < 6; i++ {
		game.TestFields = append(game.TestFields, "")
	}

	//bot snake
	game.createSnakes()

	for i := 0; i < game.FoodNumber; i++ {
		game.setNewFoodPosition()
	}

	return game
}

func (game *Game) Run(directionChan1 chan int, directionChan2 chan int, directionChanBot1 chan int, runBotCalcChan1 chan bool, directionChanBot2 chan int, runBotCalcChan2 chan bool) {
	ticker := time.NewTicker(game.Speed)
	defer ticker.Stop()

	for {
		select {
		case newDirection0 := <-directionChan1:
			if game.shouldUpdateDirection(game.Snakes[0].Direction, newDirection0) {
				game.mu.Lock()
				game.Snakes[0].Direction = newDirection0
				// game.TestField1 = fmt.Sprintln("nor dir: ", newDirection0)
				game.mu.Unlock()
			}
		// case newDirection1 := <-directionChan2:
		// 	if game.shouldUpdateDirection(game.Snakes[1].Direction, newDirection1) {
		// 		game.mu.Lock()
		// 		game.Snakes[1].Direction = newDirection1
		// 		game.mu.Unlock()
		// 	}

		case newDirectionBot := <-directionChanBot1:
			game.mu.Lock()
			game.Snakes[0].Direction = newDirectionBot
			game.mu.Unlock()

		case newDirectionBot := <-directionChanBot2:
			game.mu.Lock()
			game.Snakes[1].Direction = newDirectionBot
			game.mu.Unlock()

		case <-ticker.C:
			if game.shouldContinue() {
				game.updateItemState()
				runBotCalcChan1 <- true
				// runBotCalcChan2 <- true
			}
			game.updateScreen()
		}
	}
}

func (game *Game) Run2(playerDirChan []chan int, botDirChans []chan int, botRunChanes []chan bool) {
	ticker := time.NewTicker(game.Speed)
	defer ticker.Stop()

	cases := make([]reflect.SelectCase, len(playerDirChan)+len(botDirChans)+1)
	//PLAYER CHANs
	if len(playerDirChan) > 0 {
		for i, ch := range playerDirChan {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}
	}
	//BOT CHANs
	for i, ch := range botDirChans {
		j := i + len(playerDirChan)
		cases[j] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	//TIMER CHAN
	tc := make(chan time.Time, 1)
	go func() {
		for {
			if !game.IsOver && !game.isPaused() {
				tc <- <-ticker.C
			}
		}
	}()
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tc)}

	// game.TestField1 = fmt.Sprintf("len %v", len(cases))

	for {
		chosen, value, _ := reflect.Select(cases)
		game.TestFieldError = fmt.Sprintf("Read from channel %#v and received %v\n", chosen, value)

		if chosen != len(cases)-1 {
			if game.shouldUpdateDirection(game.Snakes[chosen].Direction, value.Interface().(int)) {
				game.mu.Lock()
				game.Snakes[chosen].Direction = value.Interface().(int)
				game.mu.Unlock()
			}
		} else {
			if game.shouldContinue() {
				game.updateItemState()
				for _, v := range botRunChanes {
					v <- true
				}
			}
			game.updateScreen()
		}
	}
}

func newFood(x int, y int) Food {
	var food Food
	//Ascii A-Z
	// minCap := 65
	// maxCap := 90
	//Ascii a-z
	minNor := 96
	maxNor := 122

	// rNumber := rand.Intn(10)

	// if rNumber < 2 {
	// 	food = Food{
	// 		Coordinates: newCoordinate(x, y),
	// 		Letter:      string(rand.Intn(maxCap-minCap+1) + minCap),
	// 		Point:       5,
	// 	}
	// } else {
	food = Food{
		Coordinates: newCoordinate(x, y),
		Letter:      string(rand.Intn(maxNor-minNor+1) + minNor),
		Point:       1,
	}
	// }

	return food
}

//----------Control-----------------------------------------------------------

func (game *Game) handleKeyBoardEvents(directionChanArray []chan int) {
	defer func() {
		for i := 0; i < len(directionChanArray); i++ {
			close(directionChanArray[i])
		}
	}()

	for {
		switch event := game.Screen.PollEvent().(type) {
		case *tcell.EventResize:
			game.resizeScreen()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
				game.exit()
			}
			if !game.hasStarted() && event.Key() == tcell.KeyEnter {
				game.start()
			}
			if !game.hasEnded() {
				for i := 0; i < game.PlayerNumber; i++ {
					if string(event.Rune()) == game.settings.PlayersControlSettings[i].Left {
						directionChanArray[i] <- Left
					}
					if string(event.Rune()) == game.settings.PlayersControlSettings[i].Right {
						directionChanArray[i] <- Right
					}
					if string(event.Rune()) == game.settings.PlayersControlSettings[i].Down {
						directionChanArray[i] <- Down
					}
					if string(event.Rune()) == game.settings.PlayersControlSettings[i].Up {
						directionChanArray[i] <- Up
					}
				}
				if event.Key() == tcell.KeyBackspace {
					game.Pause()
				}
			} else {
				if event.Rune() == 'y' {
					game.reStart()
				}
				if event.Rune() == 'n' {
					game.exit()
				}
			}
		}
	}
}

func (g *Game) setNewFoodPosition() {
	var availableCoordinates []Coordinate

	for _, coordinates := range g.Board.area {
		for _, snake := range g.Snakes {
			if !snake.contains(coordinates) {
				availableCoordinates = append(availableCoordinates, coordinates)
			}
		}
	}
	foodPosition := availableCoordinates[rand.Intn(len(availableCoordinates))]
	g.Food = append(g.Food, newFood(foodPosition.x, foodPosition.y))
}

func (g *Game) createSnakes() {
	widthForOneSnake := g.Board.width / (g.PlayerNumber + g.BotNumber)
	for i := 0; i < g.PlayerNumber; i++ {
		newSnake := newSnake((i+1)*widthForOneSnake/2, false)
		g.Snakes = append(g.Snakes, newSnake)
	}

	for i := 0; i < g.BotNumber; i++ {
		newSnake := newSnake((g.PlayerNumber+i+1)*widthForOneSnake/2, true)
		g.Snakes = append(g.Snakes, newSnake)
	}
}

func (g *Game) reCreateSnakes() {
	widthForOneSnake := g.Board.width / (g.PlayerNumber + g.BotNumber)
	for i := 0; i < g.PlayerNumber; i++ {
		newSnake := newSnake((i+1)*widthForOneSnake/2, false)
		//g.Snakes = append(g.Snakes, newSnake)
		g.Snakes[i] = newSnake
	}

	for i := 0; i < g.BotNumber; i++ {
		newSnake := newSnake((g.PlayerNumber+i+1)*widthForOneSnake/2, true)
		// g.Snakes = append(g.Snakes, newSnake)
		g.Snakes[g.PlayerNumber+i] = newSnake
	}
}

func (g *Game) updateItemState() {
	for i, currentSnake := range g.Snakes {

		if currentSnake.canMove(g.Board, g.Snakes) {
			currentSnake.move()

			for _, food := range g.Food {
				if currentSnake.CanEat(&food) {
					currentSnake.eat(&food)
					g.removeAndAddFood(food)
				}
			}
		} else {
			// ez valamiért megoldja az egyszerre lépünk egy mezőre problémát
			//valószínüleg azért mert removeolja az ütközés pillanatába, de nem menti el ezt és következő ciklusba ütközés előttről folytatja
			// if len(g.Snakes) > 1 {
			// 	removeElementFromSlice(g.Snakes, g.Snakes[i])
			// } else {
			g.over(i)
			// }
		}
	}
}

func removeElementFromSlice(slice []*Snake, s *Snake) {
	// return append(s[:index], s[index+1:]...)
	newSnakes := make([]*Snake, 0)
	for _, v := range slice {
		if v != s {
			// if v.SnakeParts[0] != s.SnakeParts[0] {
			newSnakes = append(newSnakes, v)
		}
	}
}

func (g *Game) Pause() {
	if g.IsPaused {
		g.IsPaused = false
	} else {
		g.IsPaused = true
	}
}

func (g *Game) botControl(snake *Snake, botChan chan int, runBotCalcChan1 chan bool, snakeNumber int) {
	for {
		if <-runBotCalcChan1 {
			// startTime := time.Now().UnixMilli()
			botSnake := g.Snakes[snakeNumber]

			headCordinate := botSnake.SnakeParts[0].Coordinate
			foodCordinate := g.Food[len(g.Food)-1].Coordinates

			g.TestFields[snakeNumber] = fmt.Sprintf("%v Head Pos: (%v,%v) - Food (%v,%v)", snakeNumber, headCordinate.x, headCordinate.y, g.Food[len(g.Food)-1].Coordinates.x, g.Food[0].Coordinates.y)

			world := ParseSnakeWorld(g, botSnake)

			p, _, _ := Path(world.From(), world.To(), g)
			g.BotPaths[snakeNumber] = world.getPathCoordinates(p)
			// g.TestField4 = fmt.Sprintf("dist: %v, found: %v", dist, found)
			var nextstep int
			if len(world.getPathCoordinates(p)) >= 2 {
				nextCoor := world.getPathCoordinates(p)[len(world.getPathCoordinates(p))-2]
				nextstep = g.calculateDirection2(headCordinate, nextCoor, botSnake)
			} else {
				nextstep = g.calculateDirection2(headCordinate, foodCordinate, botSnake)
			}

			botChan <- nextstep
		}
	}
}

func (g *Game) calculateDirection(currentHeadPosition Coordinate, foodPosition Coordinate, snake *Snake) int {
	var difference Coordinate
	difference.x = currentHeadPosition.x - foodPosition.x
	difference.y = currentHeadPosition.y - foodPosition.y

	var dir int
	for {
		//2
		if difference.x < 0 {
			dir = Right
			//1
		} else if difference.x > 0 {
			//3
		} else if difference.y < 0 {
			dir = Down

			//0
		} else if difference.y > 0 {
			dir = Up
		} else {
			return -5
		}
		cm := snake.canMoveBot(g.Board, g.Snakes, dir)
		if cm {
			break
		}
	}
	return dir
}
func (g *Game) calculateDirection2(currentHeadPosition Coordinate, goalCoordinate Coordinate, snake *Snake) int {
	var difference Coordinate
	difference.x = currentHeadPosition.x - goalCoordinate.x
	difference.y = currentHeadPosition.y - goalCoordinate.y

	var dir int = 0
	var right bool = false
	var left bool = false
	var up bool = false
	var down bool = false
	for i := 0; i < 4; i++ {
		//2
		if (difference.x < 0 || left) && !right {
			dir = Right
			right = true
			//1
		} else if (difference.x > 0 || right) && !left {
			dir = Left
			left = true
			//3
		} else if (difference.y < 0 || up) && !down {
			dir = Down
			down = true
			//0
		} else if (difference.y > 0 || down) && !up {
			dir = Up
			up = true
		} else {
			return -5
		}
		cm := snake.canMoveBot(g.Board, g.Snakes, dir)
		if cm {
			break
		}
	}
	return dir
}

// -----Display------------------------------------------------------------------------------
// Display the game board.
func (g *Game) drawBoard() {
	width, height := g.Board.width, g.Board.height

	boardStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	g.Screen.SetContent(0, 0, tcell.RuneCkBoard, nil, boardStyle)
	for i := 1; i < width; i++ {
		g.Screen.SetContent(i, 0, tcell.RuneCkBoard, nil, boardStyle)
	}
	g.Screen.SetContent(width, 0, tcell.RuneCkBoard, nil, boardStyle)

	for i := 1; i < height; i++ {
		g.Screen.SetContent(0, i, tcell.RuneCkBoard, nil, boardStyle)
	}

	g.Screen.SetContent(0, height, tcell.RuneCkBoard, nil, boardStyle)

	for i := 1; i < height; i++ {
		g.Screen.SetContent(width, i, tcell.RuneCkBoard, nil, boardStyle)
	}

	g.Screen.SetContent(width, height, tcell.RuneCkBoard, nil, boardStyle)

	for i := 1; i < width; i++ {
		g.Screen.SetContent(i, height, tcell.RuneCkBoard, nil, boardStyle)
	}

	fullWidth, fullHeight := g.Screen.Size()
	// g.drawText(1, height+1, width, height+10, fmt.Sprintf("P1 Score:%d", g.Snakes[0].Score))
	// g.drawText(1, height+2, width, height+10, fmt.Sprintf("P2 Score:%d", g.Snakes[1].Score))
	textHeight := height + 1
	score := ""
	for i := 0; i < len(g.Snakes); i++ {
		score += fmt.Sprintf("P%v score %v - ", i+1, g.Snakes[i].Score)
	}
	// g.drawText(1, textHeight, width, height+10, fmt.Sprintf("P1 Score:%d", g.Snakes[i].Score))
	g.drawText(1, textHeight, fullWidth, fullHeight, fmt.Sprintf("%v", score))
	textHeight++
	// g.drawText(1, textHeight, width, height+10, "Press ESC or Ctrl+C to quit")
	// textHeight++
	// g.drawText(1, textHeight, width, height+10, "Press arrow keys to control direction")
	// textHeight++

	for _, v := range g.TestFields {
		g.drawText(1, textHeight, fullWidth, fullHeight, v)
		textHeight++
	}

	textHeight++
	for key, element := range g.BotPaths {
		g.drawText(1, textHeight, fullWidth-5, fullHeight, fmt.Sprintf("%v Snake path: %v", key, element))
		textHeight++
		textHeight++
		textHeight++
		g.drawText(1, textHeight, fullWidth, fullHeight, fmt.Sprintf("---"))
		textHeight++
	}
}

// Display text in terminal.
func (g *Game) drawText(x1, y1, x2, y2 int, text string) {
	row := y1
	col := x1
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for _, r := range text {
		g.Screen.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func (g *Game) drawSnake() {
	for j, currentSnake := range g.Snakes {
		var a tcell.Color
		if !currentSnake.IsBot {
			a = tcell.Color(tcell.ColorNames[g.settings.PlayersControlSettings[j].Color])
		} else {
			a = tcell.Color((j + 2) * 10)
		}
		snakeStyle := tcell.StyleDefault.Background(a)
		for i, part := range currentSnake.SnakeParts {
			if i == 0 {
				g.Screen.SetContent(part.Coordinate.x, part.Coordinate.y, []rune(part.Letter)[0], nil, snakeStyle) //tcell.RuneBullet
			} else {
				g.Screen.SetContent(part.Coordinate.x, part.Coordinate.y, []rune(part.Letter)[0], nil, snakeStyle) //tcell.RuneCkBoard
			}
		}
	}
}

func (g *Game) drawLoading() {
	if !g.hasStarted() {
		g.drawText(g.Board.width/2-12, g.Board.height/2, g.Board.width/2+13, g.Board.height/2, fmt.Sprintf("Press <ENTER> To Continue"))
	}
}

func (g *Game) drawFood() {
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)
	for _, food := range g.Food {
		g.Screen.SetContent(food.Coordinates.x, food.Coordinates.y, []rune(food.Letter)[0], nil, style)
	}

}

func (g *Game) drawEnding() {
	if g.hasEnded() && g.hasStarted() {
		g.drawText(g.Board.width/2-5, g.Board.height/2-1, g.Board.width/2+10, g.Board.height/2, fmt.Sprintf("Game over P%v lost", g.whoLost+1))
		g.drawText(g.Board.width/2-5, g.Board.height/2, g.Board.width/2+10, g.Board.height/2, "New Game? y/n")
	}
}

func (g *Game) drawPause() {
	if g.isPaused() {
		g.drawText(g.Board.width/2-5, g.Board.height/2, g.Board.width/2+10, g.Board.height/2, "PAUSED")
	}
}

func (g *Game) updateScreen() {
	g.Screen.Clear()
	g.drawBoard()
	g.drawSnake()
	g.drawFood()

	g.drawLoading()
	g.drawEnding()
	g.drawPause()

	g.Screen.Show()
}

//-----Get/Set-------------------------------------------------------------------

func (g *Game) resizeScreen() {
	g.mu.Lock()
	g.Screen.Sync()
	g.mu.Unlock()
}

func (g *Game) exit() {
	g.mu.Lock()
	g.Screen.Fini()
	g.mu.Unlock()
	os.Exit(0)
}

func (g *Game) start() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsStart = true
	g.IsPaused = false
	g.IsOver = false
}

func (g *Game) reStart() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsStart = false
	g.IsOver = false
	g.reCreateSnakes()
}

func (g *Game) hasStarted() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.IsStart
}
func (g *Game) hasEnded() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.IsOver
}

func (g *Game) isPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.IsPaused
}

func (g *Game) shouldContinue() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return !g.IsOver && g.IsStart && !g.IsPaused
}

func (g *Game) over(i int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsOver = true
	g.whoLost = i
}

func (g *Game) shouldUpdateDirection(currentDirection int, direction int) bool {
	if currentDirection == direction {
		return false
	}
	if currentDirection == Right && direction != Left {
		return true
	}
	if currentDirection == Left && direction != Right {
		return true
	}
	if currentDirection == Up && direction != Down {
		return true
	}
	if currentDirection == Down && direction != Up {
		return true
	}
	return false
}

func (g *Game) removeFood(food Food) {
	newFoodList := make([]Food, 0)
	for _, value := range g.Food {
		b := value.Coordinates.x == food.Coordinates.x && value.Coordinates.y == food.Coordinates.y
		if !b {
			newFoodList = append(newFoodList, value)
		}
	}
	g.Food = newFoodList
}

func (g *Game) removeAndAddFood(food Food) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.setNewFoodPosition()
	g.removeFood(food)
}

//------------------------------------
