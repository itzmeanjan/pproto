package main

import (
	"log"
	"time"
)

func main() {
	file := "data.bin"
	count := 1000

	start := time.Now()
	ret := WriteAllToFile(file, count)
	end := time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Sequential ]\n", count, end.Sub(start))
	}

	start = time.Now()
	ret = ConcurrentWriteAllToFile(file, count)
	end = time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Concurrent ]\n", count, end.Sub(start))
	}

}
