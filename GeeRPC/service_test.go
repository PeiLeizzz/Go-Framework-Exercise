package geerpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sum2(args *Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	_assert(len(s.method) == 2, "wrong service Method, expect 2, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")

	argv4 := Args{Num1: 1, Num2: 3}
	replyv = mType.newReplyv()
	err = s.call(mType, reflect.ValueOf(argv4), replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 2, "failed to call Foo.Sum")

	// 如果 Sum 中第一个参数类型是 *Args
	// 那么直接 argv.Set() 会 Panic
	// 因为在 newArgv 返回的是 reflect.New(m.ArgType.Elem()) 不可寻址
	// 可以改成 argv.Elem.Set(...)
	// 在正常 main 中使用时，不用担心指针/值类型 Set 有区别的问题
	// 在 server 接收时，只需要拿 newArgv() 初始化后的 argv 去放心接收即可
	// 因为 JSON/GOB 会帮我们做好赋值的问题
	mType = s.method["Sum2"]

	argv2 := mType.newArgv()
	replyv = mType.newReplyv()
	argv2.Elem().Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err = s.call(mType, argv2, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum2")

	argv3 := &Args{Num1: 1, Num2: 3}
	replyv = mType.newReplyv()
	err = s.call(mType, reflect.ValueOf(argv3), replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 2, "failed to call Foo.Sum2")
}
