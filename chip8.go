package main

import (
	"fmt"
	"os"
	"time"
)

//https://tobiasvl.github.io/blog/write-a-chip-8-emulator/

func main() {
	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("USAGE: ./chip8 [input rom]")
		return
	}
	ticker := time.NewTicker(1 * time.Microsecond)
	fileBytes := make([]byte, 4096)
	file.Read(fileBytes)
	cpu := InitCpu()
	cpu.LoadROM(fileBytes)
	cpu.display.DisableDebug()
	go func() {
		for {
			if cpu.display.debug.enabled {
				if cpu.display.debug.step {
					cpu.CpuCycle()
					time.Sleep(20 * 1000 * 1000)
				}
			} else {
				WaitForPulse(ticker.C)
				cpu.CpuCycle()
			}
		}
	}()
	RunDisplay(cpu.display)

}

func WaitForPulse(channel <-chan time.Time) {
	for {
		if len(channel) > 0 {
			<-channel
			break
		}
	}
}
