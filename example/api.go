package main

import (
	"github.com/tiaotiao/web"
)

type Api struct {
	MessageMgr *MessageManager
}

func NewApi() *Api {
	a := new(Api)
	a.MessageMgr = NewMessageManager()
	return a
}

func (a *Api) PostMessage(c *web.Context) interface{} {
	args := struct {
		Message string `web:"message,required"` // required
		Remark  string `web:"remark"`           // not required
	}{}

	err := web.Scheme(c.Values, &args) // scheme args manually
	if err != nil {
		return err // bad requrest error
	}

	msg := a.MessageMgr.Add(args.Message, args.Remark)
	if msg == nil {
		return "failed" // return a string
	}

	return msg // return a struct
}

func (a *Api) GetMessage(c *web.Context, args struct {
	Id int64 `web:"id,required"`
}) interface{} { // scheme args automatically

	msg := a.MessageMgr.Get(args.Id)
	if msg == nil {
		return web.NewError("msg not found", web.StatusNotFound) // new error
	}
	return msg
}

func (a *Api) GetMessages(c *web.Context, args struct {
	Limit int `web:"limit,20"` // with default value
}) interface{} {

	msgs := a.MessageMgr.List()

	total := len(msgs)
	if total > args.Limit {
		msgs = msgs[:args.Limit]
	}

	return web.Result{"msgs": msgs, "total": total} // return a map
}
