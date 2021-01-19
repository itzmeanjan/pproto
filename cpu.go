package main

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"

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

	data, err := proto.Marshal(cpu)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return nil
	}

	return data

}

// WriteCPUDataToFile - Receives binary data to be written to file over
// go channel and writes that along with respective size of data
//
// Writing size is important because while deserializing we'll require
// that
func WriteCPUDataToFile(fd io.Writer, data chan []byte, stop chan bool) bool {

	for {
		select {

		case <-stop:
			break
		case d := <-data:

			// store size of message ( in bytes ), in a byte array first
			// then that's to be written on file handle
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(len(d)))

			// first write size of proto message in 4 byte space
			if _, err := fd.Write(buf); err != nil {

				log.Printf("[!] Error : %s\n", err.Error())
				return false

			}

			// then write actual message
			if _, err := fd.Write(d); err != nil {

				log.Printf("[!] Error : %s\n", err.Error())
				return false

			}

		}
	}

	log.Println("[+] Writing to file completed")
	return true

}

// WriteAllToFile - Generate random CPU data `count` times
// and store them in data file provided
func WriteAllToFile(file string, count int) bool {

	// truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	for i := 0; i < count; i++ {

		if !WriteCPUDataToFile(fd, nil) {
			return false
		}

	}

	return true

}

// ConcurrentWriteAllToFile - Concurrently generate random CPU data `count` times
// using worker pool and write them in data file provided
//
// Nothing but concurrent implementation of above function
func ConcurrentWriteAllToFile(file string, count int) bool {

	// truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	// creating worker pool of size same as number of CPUs
	// available on machine
	pool := wp.New(runtime.NumCPU())

	// lock to be used for synchronization among
	// multiple competing workers
	var lock sync.Mutex

	for i := 0; i < count; i++ {

		// submitting job to pool
		pool.Submit(func() {

			WriteCPUDataToFile(fd, &lock)

		})

	}

	// waiting for all submitted jobs to get completed
	pool.StopWait()

	return true

}
