package kernel

import (
	"delivery/internal/pkg/mathx"
	"errors"
	"math/rand"
)

const (
	minX = 1
	minY = 1
	maxX = 10
	maxY = 10
)

var (
	ErrInvalidCoordinateX = errors.New("invalid coordinate x")
	ErrInvalidCoordinateY = errors.New("invalid coordinate y")
)

type Location struct {
	x int
	y int
}

func NewLocation(x, y int) (*Location, error) {
	if x < minX || x > maxX {
		return nil, ErrInvalidCoordinateX
	}
	if y < minY || y > maxY {
		return nil, ErrInvalidCoordinateY
	}
	return &Location{x: x, y: y}, nil
}

func MustNewLocation(x, y int) Location {
	location, err := NewLocation(x, y)
	if err != nil {
		panic(err)
	}
	return *location
}

func (l Location) X() int { return l.x }

func (l Location) Y() int { return l.y }

func RandomLocation() *Location {
	var loc, _ = NewLocation(rand.Intn(maxX)+1, rand.Intn(maxY)+1)
	return loc
}

func (l *Location) Equals(o *Location) bool {
	return l != nil && o != nil && l.x == o.x && l.y == o.y
}

func (l *Location) Distance(o *Location) int {
	return mathx.Abs(l.x-o.x) + mathx.Abs(l.y-o.y)
}

func (l *Location) IsEmpty() bool {
	return l.x == 0 && l.y == 0
}
