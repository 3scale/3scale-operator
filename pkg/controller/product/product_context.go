package product

import (
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type Context struct {
	product *threescaleapi.Service
}

func (c *Context) SetProduct(p *threescaleapi.Service) {
	c.product = p
}

func (c *Context) Product() *threescaleapi.Service {
	return c.product
}
