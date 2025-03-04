package audio

// #include <stdlib.h>
// #include <stdio.h>
// #include <stdint.h>
// #include <speechapi_c_common.h>
// #include <speechapi_c_audio_config.h>
//
// typedef struct {
//     void *context;
//     int32_t (*read)(void *context, uint8_t *buffer, uint32_t size);
//     void (*close)(void *context);
// } spx_go_stream_callbacks;
//
// typedef struct {
//     spx_go_stream_callbacks callbacks;
// } spx_go_stream;
import "C"
import (
	"io"
	"sync"
	"unsafe"
)

type InMemoryAudioStream struct {
    data    []byte
    position int
}

func NewInMemoryAudioStream(data []byte) *InMemoryAudioStream {
    return &InMemoryAudioStream{data: data, position: 0}
}

func (s *InMemoryAudioStream) Read(buffer []byte) (int, error) {
    if s.position >= len(s.data) {
        return 0, io.EOF
    }
    bytesToRead := len(buffer)
    if s.position+bytesToRead > len(s.data) {
        bytesToRead = len(s.data) - s.position
    }
    copy(buffer, s.data[s.position:s.position+bytesToRead])
    s.position += bytesToRead
    return bytesToRead, nil
}

func (s *InMemoryAudioStream) Close() {
    s.data = nil // 释放内存
    s.position = 0
}

var streamIdMutex sync.Mutex
var streamId int64

func getNextStreamId() int64 {
    streamIdMutex.Lock()
    defer streamIdMutex.Unlock()
    streamId++
    return streamId
}

func (s *InMemoryAudioStream) getHandle() C.SPXHANDLE {
    streamId := getNextStreamId()
    cbs := C.spx_go_stream_callbacks{
        read:  C.go_stream_read,
        close: C.go_stream_close,
        context: unsafe.Pointer(s),
    }
    stream := C.spx_go_stream{
        callbacks: cbs,
    }
    return C.SPXHANDLE(streamId)
}

//export go_stream_read
func go_stream_read(context unsafe.Pointer, buffer *C.uint8_t, size C.uint32_t) C.int32_t {
    stream := (*InMemoryAudioStream)(context)
    buf := C.GoBytes(unsafe.Pointer(buffer), C.int(size))
    n, err := stream.Read(buf)
    if err != nil && err != io.EOF {
        println("read error", err.Error())
        return C.int32_t(-1)
    }
    return C.int32_t(n)
}

//export go_stream_close
func go_stream_close(context unsafe.Pointer) {
    stream := (*InMemoryAudioStream)(context)
    err := stream.Close()
    if err != nil {
        println("close error", err.Error())
    }
    return
}