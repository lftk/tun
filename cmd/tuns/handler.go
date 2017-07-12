package main

import (
	"github.com/gin-gonic/gin"
)

type handler struct {
	storage
}

func (h *handler) CreateProxy(ctx gin.Context) {
	var (
		err  error
		data proxy
	)

	err = ctx.Bind(&data)
	if err != nil {
		return
	}

	err = h.storage.Store(&data)
	if err != nil {
		return
	}
}

func (h *handler) DeleteProxy(ctx gin.Context) {
	var (
		err  error
		data struct {
			Name string
		}
	)

	err = ctx.Bind(&data)
	if err != nil {
		return
	}

	err = h.storage.Delete(data.Name)
	if err != nil {
		return
	}
}

func (h *handler) UpdateProxy(ctx gin.Context) {
	var (
		err  error
		data proxy
	)

	err = ctx.Bind(&data)
	if err != nil {
		return
	}

	err = h.storage.Update(&data)
	if err != nil {
		return
	}
}
