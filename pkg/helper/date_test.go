package helper

import (
	"fmt"
	"testing"
)

func TestNow(t *testing.T) {
	fmt.Printf("Current time: %s\n", Now())
}

func TestNowUTC(t *testing.T) {
	fmt.Printf("Current UTC time: %s\n", NowUTC())
}

func TestNowJakarta(t *testing.T) {
	fmt.Println("Current Jakarta time:", NowJakarta())
}
