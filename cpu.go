package main

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"os"

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

	defer func() {
		control <- true
	}()

	for {
		select {

		case <-control:
			break
		case d := <-data:

			// store size of message ( in bytes ), in a byte array first
			// then that's to be written on file handle
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(len(d)))

			// first write size of proto message in 4 byte space
			if _, err := fd.Write(buf); err != nil {

				log.Printf("[!] Error : %s\n", err.Error())
				break

			}

			// then write actual message
			if _, err := fd.Write(d); err != nil {

				log.Printf("[!] Error : %s\n", err.Error())
				break

			}

		}
	}

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

	data := make(chan []byte, count)
	done := make(chan bool, count)

	for i := 0; i < count; i++ {

		go func() {

			d := Serialize(NewCPU())
			if d == nil {
				done <- false
				return
			}

			data <- d
			done <- true

		}()

	}

	control := make(chan bool)
	go WriteCPUDataToFile(fd, data, control)

	_count := 0
	for d := range done {

		if !d {
			return false
		}

		_count++

		if _count == count {
			break
		}

	}

	control <- true
	<-control

	return true

}
