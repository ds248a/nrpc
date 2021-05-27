package util

import (
	"runtime/debug"
	"unsafe"

	acodec "github.com/ds248a/nrpc/codec"
	"github.com/ds248a/nrpc/log"
)

type Empty struct{}

func Recover() {
	if err := recover(); err != nil {
		log.Error("runtime error: %v\ntraceback:\n%v\n", err, string(debug.Stack()))
	}
}

// Wraps a function-calling with panic recovery
func Safe(call func()) {
	defer Recover()
	call()
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ValueToBytes(codec acodec.Codec, v interface{}) []byte {
	if v == nil {
		return nil
	}
	var (
		err  error
		data []byte
	)
	switch vt := v.(type) {
	case []byte:
		data = vt
	case *[]byte:
		data = *vt
	case string:
		data = StrToBytes(vt)
	case *string:
		data = StrToBytes(*vt)
	case error:
		data = StrToBytes(vt.Error())
	case *error:
		data = StrToBytes((*vt).Error())
	default:
		if codec == nil {
			codec = acodec.DefaultCodec
		}
		data, err = codec.Marshal(vt)
		if err != nil {
			log.Error("ValueToBytes: %v", err)
		}
	}

	return data
}
