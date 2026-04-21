// Package delta 提供基于双向链表的增量数据缓冲区，用于测点数据的批量推送。
package delta

import (
	"dac/entity/consts"
	"dac/entity/model/rt"
)

// valueType 链表节点值类型，即测点数组
type valueType = rt.Points

// Element 双向链表的节点类型
type Element struct {
	next, prev *Element  // 前后指针
	list       *List     // 所属链表
	Value      valueType // 节点值
}

// Next 返回下一个节点，若到达末尾则返回nil
func (e *Element) Next() *Element {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev 返回上一个节点，若到达头部则返回nil
func (e *Element) Prev() *Element {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// List 固定最大长度的双向链表，超出时自动淘汰最旧数据
type List struct {
	root   Element // 哨兵节点
	len    int     // 当前长度
	maxLen int     // 最大长度
}

// Init 初始化链表
func (l *List) Init() *List {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New 创建指定最大长度的链表
func New(maxLen uint) *List {
	if maxLen == 0 {
		return nil
	}
	l := &List{
		maxLen: int(maxLen),
	}
	return l.Init()
}

// Len 返回链表当前长度
func (l *List) Len() int { return l.len }

// Front 返回链表头节点
func (l *List) Front() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back 返回链表尾节点
func (l *List) Back() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit 延迟初始化链表
func (l *List) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert 在指定节点后插入新节点
func (l *List) insert(e, at *Element) *Element {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue 在指定节点后插入值
func (l *List) insertValue(p valueType, at *Element) *Element {
	return l.insert(&Element{Value: p}, at)
}

// remove 从链表中移除节点
func (l *List) remove(e *Element) *Element {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
	return e
}

// Remove 移除指定节点并返回其值
func (l *List) Remove(e *Element) valueType {
	if e.list == l {
		l.remove(e)
	}
	return e.Value
}

// PushBack 在链表尾部追加数据，超出最大长度时淘汰头部
func (l *List) PushBack(p valueType) *Element {
	l.lazyInit()
	if l.len == l.maxLen {
		l.RemoveFront()
	}
	return l.insertValue(p, l.root.prev)
}

// RemoveFront 移除并返回链表头部数据
func (l *List) RemoveFront() valueType {
	e := l.Front()
	if e == nil {
		return nil
	}
	return l.remove(e).Value
}

// newPush 创建新的测点批次并追加到链表尾部
func (l *List) newPush(p *rt.Point) {
	v := make([]rt.Point, 0, consts.PointNumberPerMessage)
	v = append(v, *p)
	l.PushBack(v)
	return
}

// PushPoint 追加单个测点到链表，自动合并到最后一个批次
func (l *List) PushPoint(p *rt.Point) {
	if l == nil || p == nil {
		return
	}
	if l.len == 0 {
		l.newPush(p)
		return
	}
	e := l.Back()
	v := e.Value
	if len(v) < consts.PointNumberPerMessage {
		e.Value = append(e.Value, *p)
		return
	}
	l.newPush(p)
}
