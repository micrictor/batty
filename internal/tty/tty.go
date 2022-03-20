package tty

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type TTY struct {
	Path   string
	Handle *os.File
}

type HookFn func(inputCharacter rune) []byte

func New(path string) (*TTY, error) {
	ttyFile, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	return &TTY{
		Path:   path,
		Handle: ttyFile,
	}, err
}

func (t *TTY) Close() {
	t.Handle.Close()
}

func (t *TTY) Hook(hookFn HookFn) {
	queueChannel := make(chan rune, 100)
	var lastWasBackspace bool
	pollerFn := func() {
		reader := bufio.NewReader(t.Handle)
		for {
			b, err := reader.ReadByte()
			// If our buffer is full, throw out characters in the name of performance
			if cap(queueChannel) == 0 {
				continue
			}
			// Avoid hooking our own input
			if lastWasBackspace {
				lastWasBackspace = false
				continue
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Panicf("reading failed: %v", err)
			}
			if b == byte(8) {
				lastWasBackspace = true
				continue
			}
			queueChannel <- rune(b)
		}
	}
	workerFn := func() {
		for inputCharacter := range queueChannel {
			bytesToWrite := hookFn(inputCharacter)
			go t.writeToTty(bytesToWrite, make(chan error))
		}
	}
	for i := 0; i < 5; i++ {
		go pollerFn()
		go workerFn()
	}
}

func (t *TTY) writeToTty(bytesToWrite []byte, errorChan chan error) {
	// Very small sleep to give the character time to initially print
	// Based on the default windows key repeat speed of 31ms, giving ourselves a 5ms buffer
	time.Sleep(time.Millisecond * 26)
	for _, b := range bytesToWrite {
		_, _, errNo := syscall.RawSyscall(
			syscall.SYS_IOCTL,
			t.Handle.Fd(),
			syscall.TIOCSTI,
			uintptr(unsafe.Pointer(&b)),
		)
		if errNo != 0 {
			log.Printf("%v", errNo)
			errorChan <- errNo
			return
		}
	}
}
