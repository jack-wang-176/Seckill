#!/bin/bash
cat << 'INNER' > fix.go
package main

import (
    "io/ioutil"
    "strings"
	"fmt"
)

func processFile(path string, targetStr, replacement string) {
    bytes, err := ioutil.ReadFile(path)
    if err != nil { panic(err) }
    content := string(bytes)
    content = strings.Replace(content, targetStr, replacement, 1)
    ioutil.WriteFile(path, []byte(content), 0644)
}

func main() {
    // 1. user consumer
    userOld := `		for d := range msgs {
			ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)
						ctx, span := tracer.Start(ctx, "ConsumeUserRegister")
						defer span.End()
						var msg mq.UserMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				u.Logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}
			err := u.MR.RegisterUser(msg)
			if err != nil {
				if err.Error() == "username already exists" {
					u.Logger.Warn("注册幂等，用户名已存在", zap.String("username", msg.Username))
					d.Ack(false)
				} else {
					u.Logger.Error("用户注册失败", zap.Error(err))
					d.Nack(false, true)
				}
			} else {
				u.Logger.Info("用户注册成功", zap.String("username", msg.Username))
				d.Ack(false)
			}
		}`
		
	userNew := `		for d := range msgs {
			go func(d amqp091.Delivery) {
				ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)
				ctx, span := tracer.Start(ctx, "ConsumeUserRegister")
				defer span.End()
				var msg mq.UserMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					u.Logger.Error("fail to parse message", zap.Error(err))
					d.Nack(false, false)
					return
				}
				err := u.MR.RegisterUser(msg)
				if err != nil {
					if err.Error() == "username already exists" {
						u.Logger.Warn("注册幂等，用户名已存在", zap.String("username", msg.Username))
						d.Ack(false)
					} else {
						u.Logger.Error("用户注册失败", zap.Error(err))
						d.Nack(false, true)
					}
				} else {
					u.Logger.Info("用户注册成功", zap.String("username", msg.Username))
					d.Ack(false)
				}
			}(d)
		}`

    // We will do it with pure go script using replace logic if we want, or better: just rewrite the functions using simpler shell replace.
}
INNER
