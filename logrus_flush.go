package formatter

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	asyncHooks     []*AsyncHook
	asyncHookMutex sync.Mutex
)

// RegisterAsyncHook 注册一个异步Hook以便程序退出时刷新
func RegisterAsyncHook(hook *AsyncHook) {
	asyncHookMutex.Lock()
	defer asyncHookMutex.Unlock()
	asyncHooks = append(asyncHooks, hook)
}

// FlushAsyncHooks 刷新所有异步Hook中的日志
func FlushAsyncHooks() {
	asyncHookMutex.Lock()
	defer asyncHookMutex.Unlock()

	for _, hook := range asyncHooks {
		hook.Flush()
	}
}

// SetupSignalHandler 设置信号处理程序，在程序退出前刷新日志
func SetupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		fmt.Printf("Received signal %v, flushing logs before exit\n", sig)
		FlushAsyncHooks()
		os.Exit(1)
	}()
}

// init 在导入包时自动设置信号处理
func init() {
	SetupSignalHandler()
}
