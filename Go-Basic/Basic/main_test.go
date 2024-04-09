package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"testing"
)

func Perimeter(rectangle Rectangle) float64 {
	return 2 * (rectangle.Width + rectangle.Height)
}
func Area(rectangle Rectangle) float64 {
	return rectangle.Width * rectangle.Height
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}
func TestPerimeter(t *testing.T) {
	rectangle := Rectangle{10.0, 10.0}
	got := Perimeter(rectangle)

	want := 40.0
	if got != want {
		t.Errorf("got %.2f want %.2f", got, want)
	}
}

func TestArea(t *testing.T) {
	str := "5Yqg5YWlR1ZB5Lqk5rWB576k"
	decodeBytes, err := base64.StdEncoding.DecodeString(str)
	fmt.Println(decodeBytes, err)
}

type Rectangle struct {
	Width  float64
	Height float64
}

type Circle struct {
	Radius float64
}
