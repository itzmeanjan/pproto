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
		log.Printf("[!] Error : %s\n", err.Error())
		return nil
	}

	return data

}

// WriteCPUDataToFile - Create new CPU struct, serialize it
// to binary format, which is to be written file, along with it's
// size in bytes, before actual CPU data, which will help us in decoding so
func WriteCPUDataToFile(fd io.Writer, lock *sync.Mutex) bool {

	// create new message
	cpu := NewCPU()
	// serialize message in byte array form
	data := Serialize(cpu)
	if data == nil {
		return false
	}

	// store size of message ( in bytes ), in a byte array first
	// then that's to be written on file handle
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(data)))

	// If it's concurrent call, this critical section of code
	// needs to be protected using locking mechanism
	if lock != nil {

		lock.Lock()
		defer lock.Unlock()

	}

	// first write size of proto message in 4 byte space
	if _, err := fd.Write(buf); err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		return false

	}

	// then write actual message
	if _, err := fd.Write(data); err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		return false

	}

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
		WriteCPUDataToFile(fd, nil)
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
