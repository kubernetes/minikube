package controllers

import "github.com/revel/revel"
import "time"

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	go func() {
		time.Sleep(5 * time.Second)
		panic("hello!")
	}()

	s := make([]string, 0)
	revel.INFO.Print(s[0])
	return c.Render()
}
