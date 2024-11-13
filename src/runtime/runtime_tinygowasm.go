//go:build tinygo.wasm && !wasm_unknown && !wasip2

// This file is for wasm/wasip1 and for wasm/js, which both use much of the
// WASIp1 API.

package runtime

import (
	"unsafe"
)

// Implements __wasi_iovec_t.
type __wasi_iovec_t struct {
	buf    unsafe.Pointer
	bufLen uint
}

//go:wasmimport wasi_snapshot_preview1 fd_write
func fd_write(id uint32, iovs *__wasi_iovec_t, iovs_len uint, nwritten *uint) (errno uint)

// See:
// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-proc_exitrval-exitcode
//
//go:wasmimport wasi_snapshot_preview1 proc_exit
func proc_exit(exitcode uint32)

// Flush stdio on exit.
//
//export __stdio_exit
func __stdio_exit()

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		// Read the number of args (argc) and the buffer size required to store
		// all these args (argv).
		var argc, argv_buf_size uint32
		args_sizes_get(&argc, &argv_buf_size)
		if argc == 0 {
			return nil
		}

		// Obtain the command line arguments
		argsSlice := make([]unsafe.Pointer, argc)
		buf := make([]byte, argv_buf_size)
		args_get(&argsSlice[0], unsafe.Pointer(&buf[0]))

		// Convert the array of C strings to an array of Go strings.
		args = make([]string, argc)
		for i, cstr := range argsSlice {
			length := strlen(cstr)
			argString := _string{
				length: length,
				ptr:    (*byte)(cstr),
			}
			args[i] = *(*string)(unsafe.Pointer(&argString))
		}
	}
	return args
}

//go:wasmimport wasi_snapshot_preview1 args_get
func args_get(argv *unsafe.Pointer, argv_buf unsafe.Pointer) (errno uint16)

//go:wasmimport wasi_snapshot_preview1 args_sizes_get
func args_sizes_get(argc *uint32, argv_buf_size *uint32) (errno uint16)

const (
	putcharBufferSize = 120
	stdout            = 1
)

// Using global variables to avoid heap allocation.
var (
	putcharBuffer        = [putcharBufferSize]byte{}
	putcharPosition uint = 0
	putcharIOVec         = __wasi_iovec_t{
		buf: unsafe.Pointer(&putcharBuffer[0]),
	}
	putcharNWritten uint
)

func putchar(c byte) {
	putcharBuffer[putcharPosition] = c
	putcharPosition++

	if c == '\n' || putcharPosition >= putcharBufferSize {
		putcharIOVec.bufLen = putcharPosition
		fd_write(stdout, &putcharIOVec, 1, &putcharNWritten)
		putcharPosition = 0
	}
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = nanotime()
	sec = mono / (1000 * 1000 * 1000)
	nsec = int32(mono - sec*(1000*1000*1000))
	return
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	// Flush stdio buffers.
	__stdio_exit()

	// Exit the program.
	proc_exit(uint32(code))
}

func mainReturnExit() {
	syscall_Exit(0)
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

func hardwareRand() (n uint64, ok bool) {
	n |= uint64(libc_arc4random())
	n |= uint64(libc_arc4random()) << 32
	return n, true
}

// uint32_t arc4random(void);
//
//export arc4random
func libc_arc4random() uint32
