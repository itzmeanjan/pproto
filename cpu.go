package main

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	wp "github.com/gammazero/workerpool"
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

// SequentialWriteToFile - Given file name and number of protocol buffer
// entries to be written to file, it'll sequentially write those many entries
// into file
func SequentialWriteToFile(file string, count int) bool {

	// truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	for i := 0; i < count; i++ {

		data := Serialize(NewCPU())
		if data == nil {
			return false
		}

		// store size of message ( in bytes ), in a byte array first
		// then that's to be written on file handle
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(len(data)))

		// first write size of proto message in 4 byte space
		if _, err := fd.Write(buf); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

		// then write actual message
		if _, err := fd.Write(data); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

	}

	return true

}

// WriteCPUDataToFile - Receives binary data to be written to file over
// go channel and writes that along with respective size of data
//
// Writing size is important because while deserializing we'll require
// that
func WriteCPUDataToFile(fd io.Writer, data chan []byte, control chan bool) {

	status := true
	defer func() {
		control <- status
	}()

	for d := range data {

		if d == nil {
			break
		}

		// store size of message ( in bytes ), in a byte array first
		// then that's to be written on file handle
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(len(d)))

		// first write size of proto message in 4 byte space
		if _, err := fd.Write(buf); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())

			status = false
			break

		}

		// then write actual message
		if _, err := fd.Write(d); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())

			status = false
			break

		}

	}

}

// ConcurrentWriteAllToFile - Concurrently generate random CPU data `count` times
// using worker pool and write them in data file provided
//
// Nothing but concurrent implementation of above function, but file writer is
// working sequentially
func ConcurrentWriteAllToFile(file string, count int) bool {

	// truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	pool := wp.New(runtime.NumCPU())

	data := make(chan []byte, count)
	control := make(chan bool)

	go WriteCPUDataToFile(fd, data, control)

	for i := 0; i < count; i++ {

		pool.Submit(func() {

			data <- Serialize(NewCPU())

		})

	}

	pool.StopWait()

	data <- nil
	<-control

	return true

}
