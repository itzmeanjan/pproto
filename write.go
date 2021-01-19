package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"runtime"

	wp "github.com/gammazero/workerpool"
)

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

// ConcurrentWriteAllToFile - Concurrently generate random CPU data `count` times
// using worker pool and write them in data file provided
//
// Nothing but concurrent implementation of above function, but file writer is
// working sequentially
func ConcurrentWriteAllToFile(file string, count int) bool {

	// Truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	pool := wp.New(runtime.NumCPU())
	data := make(chan []byte, count)

	go WriteCPUDataToFile(fd, data)

	for i := 0; i < count; i++ {

		pool.Submit(func() {

			data <- Serialize(NewCPU())

		})

	}

	pool.StopWait()
	// letting file writer go routine know no more data
	// to be sent for writing to file, it can exit now
	data <- nil

	return true

}

// WriteCPUDataToFile - Receives binary data to be written to file over
// go channel and writes that along with respective size of data
//
// Writing size is important because while deserializing we'll require
// that
func WriteCPUDataToFile(fd io.Writer, data chan []byte) {

	for d := range data {

		// As soon as nil is received we return, by this coordinator go routine denotes
		// no more data to be sent to this go routine for writing to file
		//
		// So we'll can safely out of this loop
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
			break

		}

		// then write actual message
		if _, err := fd.Write(d); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

	}

}
