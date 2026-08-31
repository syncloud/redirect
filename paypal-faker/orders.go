package main

import (
	"strconv"
	"sync"
)

type Order struct {
	Id       string
	Currency string
	Value    string
	Captured bool
}

type Orders struct {
	mutex  sync.Mutex
	orders map[string]*Order
	next   int
}

func NewOrders() *Orders {
	return &Orders{orders: map[string]*Order{}}
}

func (o *Orders) Add(currency, value string) *Order {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.next++
	order := &Order{Id: "PAYPALORDER" + strconv.Itoa(o.next), Currency: currency, Value: value}
	o.orders[order.Id] = order
	return order
}

func (o *Orders) Get(id string) *Order {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.orders[id]
}

func (o *Orders) Capture(id string) *Order {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	order := o.orders[id]
	if order != nil {
		order.Captured = true
	}
	return order
}
