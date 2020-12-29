package quacktors

type Actor interface {
	Run(ctx *Context, message Message)
}
