package concurrency

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"Go-Basic/goasync"
)

//BG:在业务中我们经常会对很多个系统进行外部调用，
//此时各系统独立部署，此时并发访问并不会对被调用系统造成过多压力，
//并且由于大部分时候操作没有先后顺序，可以支持并发获取相应的数据，
//此时如果顺序执行的话，最长的执行时间为各个系统的总和，
//在对RT要求高的场景下无法达到预期，所以需要并发进行处理。

// InterruptType 失败类型
type InterruptType int64

const (
	FastFail          InterruptType = iota + 1 // 只要有一个错误立马终止执行
	PreExecFail                                // 按加入方法顺序来判断错误，e.g. 第二个方法执行前，如果第一个方法已经出现错误，则不执行直接返回，如果第三个执行失败，则仍然会执行第二个方法
	UninterruptedType                          // 不中断执行
)

// 并发执行的参数
type concurrencyGroupParam struct {
	timeout       time.Duration // 超时时间
	interruptType InterruptType // 中断类型
	concurrency   int           // 并发数
}

var (
	ConcurrencyGroupInterruptErr   = errors.New("interrupted!")
	ConcurrencyGroupContextDoneErr = errors.New("Context Done err!")
)

func IsInterruptErr(err error) bool {
	return errors.Is(err, ConcurrencyGroupInterruptErr)
}

func IsContextDoneErr(err error) bool {
	return errors.Is(err, ConcurrencyGroupContextDoneErr)
}

type GroupParamOptFunc func(param *concurrencyGroupParam)

func WithConcurrencyGroupParamTimeoutOpt(time time.Duration) GroupParamOptFunc {
	return func(param *concurrencyGroupParam) {
		param.timeout = time
	}
}

func WithConcurrencyGroupParamInterruptTypeOpt(interruptType InterruptType) GroupParamOptFunc {
	return func(param *concurrencyGroupParam) {
		param.interruptType = interruptType
	}
}
func WithConcurrencyGroupParamConcurrency(concurrency int) GroupParamOptFunc {
	return func(param *concurrencyGroupParam) {
		param.concurrency = concurrency
	}
}

// 错误处理的实现,需要实现一下功能
// 1.顺序保存
// 2.并发安全
// 3.支持扩容
type concurrencyGroupErrs struct {
	sync.Mutex // 扩容保证并发安全
	errs       []error
}

func newConcurrencyGroupErrs() *concurrencyGroupErrs {
	return &concurrencyGroupErrs{
		errs: make([]error, 8), // 初始化8个
	}
}
func (cge *concurrencyGroupErrs) growsErrSlice(index int) {
	if len(cge.errs) > index {
		return
	}
	cge.Lock()
	defer cge.Unlock()
	// 二次检查
	if len(cge.errs) > index {
		return
	}
	// 扩容
	growLen := 0
	if len(cge.errs) < 1024 {
		growLen = len(cge.errs)
	} else {
		growLen = int(math.Ceil(float64(len(cge.errs)) * 0.25))
	}
	newErrs := make([]error, growLen)
	cge.errs = append(cge.errs, newErrs...)
}

func (cge *concurrencyGroupErrs) setErr(err error, index int) {
	cge.growsErrSlice(index)
	cge.errs[index] = err
}

// 判断是否出现过错误
func (cge *concurrencyGroupErrs) hasError() bool {
	errs := cge.errs
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

// 前面的函数是否出现错误
func (cge *concurrencyGroupErrs) beforeHasError(index int) bool {
	errs := cge.errs
	// 执行到后面时可能前面有些错误还没有加入，所以取错误的长度进行遍历
	if index > len(errs) {
		index = len(errs)
	}
	for i := 0; i < index-1; i++ {
		if errs[i] != nil {
			return true
		}
	}
	return false
}
func (cge *concurrencyGroupErrs) getAllErrs() []error {
	errs := make([]error, 0, len(cge.errs))
	for _, err := range cge.errs {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (cge *concurrencyGroupErrs) getFirstErr() error {
	for _, err := range cge.errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// 并发执行
type groupGoFunc func(ctx context.Context) error

// 执行函数的信息
type groupFuncInfo struct {
	index  int         // 顺序
	f      groupGoFunc // 执行函数
	cancel func()      // 取消函数
}

// ConcurrencyGroup 执行group的结构
type ConcurrencyGroup struct {
	cancel        func() // 取消函数
	wg            sync.WaitGroup
	concurrencyCh chan struct{} // 控制并发的channel
	groupFuncs    []*groupFuncInfo
	*concurrencyGroupParam
	*concurrencyGroupErrs
}

func NewConcurrencyGroup(opts ...GroupParamOptFunc) *ConcurrencyGroup {
	cg := &ConcurrencyGroup{
		cancel:               nil,
		wg:                   sync.WaitGroup{},
		groupFuncs:           make([]*groupFuncInfo, 0),
		concurrencyGroupErrs: newConcurrencyGroupErrs(),
		concurrencyGroupParam: &concurrencyGroupParam{
			concurrency: 8, //默认并发数为8 ，这里是指真正执行操作的并发数
		},
	}
	for _, opt := range opts {
		opt(cg.concurrencyGroupParam)
	}
	cg.concurrencyCh = make(chan struct{}, cg.concurrency)
	return cg
}

// addGoFunc
// @Description: 增加执行的函数
func (cg *ConcurrencyGroup) addGoFunc(f groupGoFunc, ctx context.Context) *groupFuncInfo {
	cg.wg.Add(1)
	goFunc := &groupFuncInfo{
		index: len(cg.groupFuncs), // 下标进行顺序索引
		f:     f,
	}
	cg.groupFuncs = append(cg.groupFuncs, goFunc)
	return goFunc
}
func (cg *ConcurrencyGroup) Wait() []error {
	cg.wg.Wait()
	// 执行退出操作
	_ = cg.Close()
	// 裁切掉过多扩容的部分
	return cg.errs[0:len(cg.groupFuncs)]
}

// isInterruptExec
// @Description: 是否中断执行
func (cg *ConcurrencyGroup) isInterruptExec(funcIndex int) bool {
	switch cg.interruptType {
	case PreExecFail:
		return cg.beforeHasError(funcIndex)
	case FastFail:
		// 有错误立即中断
		return cg.hasError()
	case UninterruptedType:
		return false
	default:
		// 默认策略是前置有执行失败则中断
		return cg.beforeHasError(funcIndex)
	}
}

func (cg *ConcurrencyGroup) doFunc(f *groupFuncInfo, ctx context.Context) chan error {
	errChan := make(chan error, 1)
	// 不执行直接返回
	if cg.isInterruptExec(f.index) {
		errChan <- ConcurrencyGroupInterruptErr
		return errChan
	}
	go func() {
		defer func() {
			r := recover()
			err := goasync.PanicErrHandler(r)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("%q", err))
			}
		}()
		// 二次检查，被调度时产生错误了
		if cg.isInterruptExec(f.index) {
			errChan <- ConcurrencyGroupInterruptErr
			return
		}
		// 读取错误
		errChan <- f.f(ctx)
	}()
	return errChan
}
func (cg *ConcurrencyGroup) Close() error {
	for _, goFunc := range cg.groupFuncs {
		if goFunc.cancel != nil {
			goFunc.cancel()
		}
	}
	if cg.cancel != nil {
		cg.cancel()
	}
	return nil
}

func (cg *ConcurrencyGroup) withContextTimeout(ctx context.Context) (context.Context, func()) {
	if cg.timeout <= 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, cg.timeout)
}

func (cg *ConcurrencyGroup) Go(f groupGoFunc, ctx context.Context) {
	goFunc := cg.addGoFunc(f, ctx)
	go func() {
		// 最后执行done,避免recover中还有收尾工作没做
		defer func() {
			// 执行完成读走缓冲区数据
			<-cg.concurrencyCh
			cg.wg.Done()
		}()

		defer func() {
			r := recover()
			_ = goasync.PanicErrHandler(r)
		}()
		// 如果已经达到最大并发数，则进入等待
		cg.concurrencyCh <- struct{}{}
		// 开始执行则设置超时
		ctx, goFunc.cancel = cg.withContextTimeout(ctx)
		errChan := cg.doFunc(goFunc, ctx)
		select {
		case err := <-errChan:
			cg.setErr(err, goFunc.index)
		case <-ctx.Done(): // context被done则直接返回
			cg.setErr(ConcurrencyGroupContextDoneErr, goFunc.index)
		}
	}()
}
