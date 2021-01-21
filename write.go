package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"runtime"

	wp "github.com/gammazero/workerpool"
	press "github.com/valyala/gozstd"
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
	done := make(chan bool)

	go WriteCPUDataToFile(fd, count, data, done)

	for i := 0; i < count; i++ {

		pool.Submit(func() {

			data <- Serialize(NewCPU())

		})

	}

	pool.StopWait()
	// Blocking call i.e. waiting for writer go routine
	// to complete its job
	<-done

	return true

}

// WriteCPUDataToFile - Receives binary data to be written to file over
// go channel and writes that along with respective size of data
//
// Writing size is important because while deserializing we'll require
// that
func WriteCPUDataToFile(fd io.Writer, count int, data chan []byte, done chan bool) {

	// Letting coordinator know writing to file has been completed
	// or some kind of error has occurred
	//
	// To be invoked when getting out of this execution scope
	defer func() {
		done <- true
	}()

	// How many data chunks received over channel
	//
	// To be compared against data chunks which were supposed
	// to be received, before deciding whether it's time to get out of
	// below loop or not
	var iter int

	for d := range data {

		// received new data which needs to be written to file
		iter++

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

		// As soon as this condition is met,
		// we can safely get out of this loop
		// i.e. denoting all processing has been done
		if iter == count {
			break
		}

	}

}

// CompressedSequentialWriteToFile - Writing `zstd` compressed content to file
// in sequential fashion
//
// Main objective is to reduce size of final snapshot data file
func CompressedSequentialWriteToFile(file string, count int) bool {

	// truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Error : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()
	var compressed []byte

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

		// before writing protocol buffer serialized data chunk
		// compressing it using `zstd`
		compressed = press.Compress(compressed[:0], data)

		// then write compressed message
		if _, err := fd.Write(compressed); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

	}

	return true

}
