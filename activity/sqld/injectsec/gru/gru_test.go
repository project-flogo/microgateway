package gru

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestSerialize(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	inputSize := 256 + len(Chunks)
	embeddingSize := 10
	outputSize := 2
	hiddenSizes := []int{5}
	a := NewModel(rnd, 2, inputSize, embeddingSize, outputSize, hiddenSizes)
	buffer := &bytes.Buffer{}
	err := a.Write(buffer)
	if err != nil {
		t.Fatal(err)
	}
	b := NewModel(rnd, 2, inputSize, embeddingSize, outputSize, hiddenSizes)
	err = b.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}
	err = a.compare(b)
	if err != nil {
		t.Fatal(err)
	}
}
