package snake

import (
	"errors"
	"log"
)

type SnakePart struct {
	Coordinate Coordinate
	Letter     string
}

type Snake struct {
	SnakeParts []SnakePart
	Direction  int
	Score      int
	IsBot      bool
}

func (s *Snake) canMove(board *Board, snakes []*Snake) bool {
	nextHeadPosition, err := s.nextHeadPosition()

	if err != nil {
		log.Fatal(err.Error())
	}

	for _, snake := range snakes {
		for _, position := range snake.SnakeParts {
			if nextHeadPosition == position.Coordinate {
				return false
			}
		}
	}

	switch s.Direction {
	case Up:
		return nextHeadPosition.y > 0
	case Left:
		return nextHeadPosition.x > 0
	case Right:
		return nextHeadPosition.x < board.width
	case Down:
		return nextHeadPosition.y < board.height
	}

	return true
}

func (s *Snake) canMoveBot(board *Board, snakes []*Snake, newDir int) bool {
	nextHeadPosition, err := s.nextHeadPositionBot(newDir)

	if err != nil {
		log.Fatal(err.Error())
	}

	for _, snake := range snakes {
		for _, position := range snake.SnakeParts {
			if nextHeadPosition == position.Coordinate {
				return false
			}
		}
	}

	switch newDir {
	case Up:
		return nextHeadPosition.y > 0
	case Left:
		return nextHeadPosition.x > 0
	case Right:
		return nextHeadPosition.x < board.width
	case Down:
		return nextHeadPosition.y < board.height
	}

	return true
}

func (s *Snake) nextHeadPosition() (Coordinate, error) {
	var head Coordinate
	var err error
	switch s.Direction {
	case Up:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x, (*s).SnakeParts[0].Coordinate.y-1)
	case Right:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x+1, (*s).SnakeParts[0].Coordinate.y)
	case Down:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x, (*s).SnakeParts[0].Coordinate.y+1)
	case Left:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x-1, (*s).SnakeParts[0].Coordinate.y)
	default:
		err = errors.New("error: invalid direction")
	}
	return head, err
}

func (s *Snake) nextHeadPositionBot(newDir int) (Coordinate, error) {
	var head Coordinate
	var err error
	switch newDir {
	case Up:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x, (*s).SnakeParts[0].Coordinate.y-1)
	case Right:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x+1, (*s).SnakeParts[0].Coordinate.y)
	case Down:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x, (*s).SnakeParts[0].Coordinate.y+1)
	case Left:
		head = newCoordinate((*s).SnakeParts[0].Coordinate.x-1, (*s).SnakeParts[0].Coordinate.y)
	default:
		err = errors.New("error: invalid direction")
	}
	return head, err
}
func testSnakeMove(newDir int, snakeTest Snake) (Coordinate, error) {
	var head Coordinate
	var err error
	switch newDir {
	case Up:
		head = newCoordinate(snakeTest.SnakeParts[0].Coordinate.x, snakeTest.SnakeParts[0].Coordinate.y-1)
	case Right:
		head = newCoordinate(snakeTest.SnakeParts[0].Coordinate.x+1, snakeTest.SnakeParts[0].Coordinate.y)
	case Down:
		head = newCoordinate(snakeTest.SnakeParts[0].Coordinate.x, snakeTest.SnakeParts[0].Coordinate.y+1)
	case Left:
		head = newCoordinate(snakeTest.SnakeParts[0].Coordinate.x-1, snakeTest.SnakeParts[0].Coordinate.y)
	default:
		err = errors.New("error: invalid direction")
	}
	return head, err
}

func (s *Snake) contains(coordinate Coordinate) bool {
	for _, body := range (*s).SnakeParts {
		if coordinate == body.Coordinate {
			return true
		}
	}
	return false
}

func (s *Snake) CanEat(food *Food) bool {
	headPosition := (*s).SnakeParts[0]
	return headPosition.Coordinate.x == food.Coordinates.x && headPosition.Coordinate.y == food.Coordinates.y
}

func (s *Snake) eat(food *Food) {
	//FIX food coordinates are the same as the head coordinatas
	coordinate := newCoordinate(food.Coordinates.x, food.Coordinates.y)
	coordinate = newCoordinate(0, 0)
	letter := food.Letter
	s.Score += food.Point
	(*s).SnakeParts = append((*s).SnakeParts, *newSnakePart(coordinate, letter))
}

func (s *Snake) move() {
	newBody := make([]SnakePart, 0)
	for i := 0; i < len((*s).SnakeParts); i++ {
		var coordinates Coordinate
		var err error
		var letter string
		if i == 0 {
			coordinates, err = s.nextHeadPosition()
			if err != nil {
				log.Fatalln(err.Error())
				return
			}
		} else {
			coordinates = newCoordinate((*s).SnakeParts[i-1].Coordinate.x, (*s).SnakeParts[i-1].Coordinate.y)
		}
		letter = (*s).SnakeParts[i].Letter
		newBody = append(newBody, *newSnakePart(coordinates, letter))
	}
	(*s).SnakeParts = newBody
}

func (s *Snake) testMove() Snake {
	var testSnake Snake
	testSnake = *s
	testSnake.move()
	return testSnake
}

func newSnakePart(coordinate Coordinate, letter string) *SnakePart {
	sp := SnakePart{
		Coordinate: coordinate,
		Letter:     letter,
	}
	return &sp
}

func newSnake(startX int, isBot bool) *Snake {
	var snake Snake
	body := make([]SnakePart, 0)

	body = append(body, *newSnakePart(newCoordinate(startX, 7), "H"))
	body = append(body, *newSnakePart(newCoordinate(startX, 8), "O"))
	body = append(body, *newSnakePart(newCoordinate(startX, 9), "O"))
	body = append(body, *newSnakePart(newCoordinate(startX, 10), "O"))
	body = append(body, *newSnakePart(newCoordinate(startX, 11), "O"))

	snake.SnakeParts = body
	snake.Direction = 0
	snake.Score = 0
	snake.IsBot = isBot
	return &snake
}

// copysnake
func (s *Snake) copySnake() Snake {
	var snake Snake
	body := make([]SnakePart, 0)

	for _, v := range s.SnakeParts {
		body = append(body, v)
	}
	snake.SnakeParts = body
	snake.Direction = s.Direction
	snake.Score = s.Score
	return snake
}
