package product

import (
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type Context struct {
	product *threescaleapi.Product
}

func (c *Context) SetProduct(p *threescaleapi.Product) {
	c.product = p
}

func (c *Context) Product() *threescaleapi.Product {
	return c.product
}
