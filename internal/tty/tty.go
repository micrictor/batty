package tty

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
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
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := t.Handle.Read(buf)
			// Avoid hooking our own input
			if lastWasBackspace {
				lastWasBackspace = false
				continue
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Panicf("reading failed: %v", err)
			}
			if n == 0 {
				continue
			}
			if buf[0] == byte(8) {
				lastWasBackspace = true
				continue
			}
			queueChannel <- rune(buf[0])
		}
	}()

	workerFn := func() {
		for inputCharacter := range queueChannel {
			bytesToWrite := hookFn(inputCharacter)
			go t.writeToTty(bytesToWrite, make(chan error))
		}
	}
	for i := 0; i < 10; i++ {
		go workerFn()
	}
}

func (t *TTY) writeToTty(bytesToWrite []byte, errorChan chan error) {
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
