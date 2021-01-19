package main

import (
	"math/rand"

	"github.com/itzmeanjan/pproto/pb"
	"google.golang.org/protobuf/proto"
)

// NewCPU - Create new random CPU instance
func NewCPU() *pb.CPU {

	return &pb.CPU{
		Brand:   "X",
		Name:    "Y",
		Cores:   uint32(rand.Intn(16)),
		Threads: uint32(rand.Intn(64)),
		MinGhz:  rand.Float64() * 5,
		MaxGhz:  rand.Float64() * 10,
	}

}

// Serialize - Given CPU struct, serializes it into byte array which
// can be stored in file
func Serialize(cpu *pb.CPU) []byte {

	data, err := proto.Marshal(cpu)
	if err != nil {
		return nil
	}

	return data

}
