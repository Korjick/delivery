package kernel

import (
	"errors"
	"fmt"
)

var (
	invalidLocation    = Location{x: 0, y: 0}
	InvalidLocationErr = errors.New("invalid location")
)

type Location struct {
	x int8
	y int8
}

func NewLocation(x int8, y int8) (Location, error) {
	if x > 10 || x < 1 || y > 10 || y < 1 {
		return invalidLocation, InvalidLocationErr
	}

	return Location{
		x: x,
		y: y,
	}, nil
}

func NewDefaultLocation() (Location, error) {
	return NewLocation(1, 1)
}

func (l Location) String() string {
	return fmt.Sprintf("(%d, %d)", l.x, l.y)
}

func (l Location) Equals(other Location) bool {
	return l.x == other.x && l.y == other.y
}

func (l Location) DistanceTo(target Location) int8 {
	dx := l.x - target.x
	if dx < 0 {
		dx = -dx
	}
	dy := l.y - target.y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (l Location) X() int8 {
	return l.x
}

func (l Location) Y() int8 {
	return l.y
}
