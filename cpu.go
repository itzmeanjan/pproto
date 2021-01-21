package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/itzmeanjan/pproto/pb"
	"google.golang.org/protobuf/proto"
)

// NewCPU - Create new CPU instance
func NewCPU() *pb.CPU {

	return &pb.CPU{
		Brand:   "X",
		Name:    "Y",
		Cores:   1,
		Threads: 2,
		MinGhz:  1.0,
		MaxGhz:  2.0,
	}

}

// NewRandomCPU - Create new random CPU instance
func NewRandomCPU() *pb.CPU {

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

	// synthetic delay creation, to denote we're getting
	// data from some busy IO resource i.e. DB, let's say
	time.Sleep(time.Duration(1) * time.Microsecond)

	data, err := proto.Marshal(cpu)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return nil
	}

	return data

}
