package main

import (
	"bytes"
	"sync"
)

type objectPool struct {
	bufPool        *sync.Pool
	fieldSlicePool *sync.Pool
}

func newBufferPool() objectPool {
	var bufPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	var fieldSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]responseField, 0, 10)
		},
	}
	return objectPool{
		bufPool:        &bufPool,
		fieldSlicePool: &fieldSlicePool,
	}
}

func (op *objectPool) getBuffer() *bytes.Buffer {
	return op.bufPool.Get().(*bytes.Buffer)
}

func (op *objectPool) returnBuffer(buf *bytes.Buffer) {
	buf.Reset()
	op.bufPool.Put(buf)
}

func (op *objectPool) getFieldSlice() []responseField {
	return op.fieldSlicePool.Get().([]responseField)
}

func (op *objectPool) returnFieldSlice(slice []responseField) {
	slice = slice[0:0]
	op.fieldSlicePool.Put(slice)
}
