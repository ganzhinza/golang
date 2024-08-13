package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	var in chan interface{}
	wg := &sync.WaitGroup{}
	for _, val := range cmds {
		out := make(chan interface{})
		wg.Add(1)
		go func(f cmd, in, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			f(in, out)
		}(val, in, out)
		in = out
	}
	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	userIDs := make(map[uint64]interface{})
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for userEmail := range in {
		wg.Add(1)
		go func(userEmail interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			switch userEmail := userEmail.(type) {
			case string:
				user := GetUser(userEmail)
				mu.Lock()
				if _, ok := userIDs[user.ID]; !ok {
					userIDs[user.ID] = struct{}{}
					out <- user
				}
				mu.Unlock()
			default:
				fmt.Printf("Not string!\n")
			}
		}(userEmail, wg)
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	usersBatch := make([]User, GetMessagesMaxUsersBatch)
	countInBatch := 0
	wg := &sync.WaitGroup{}
	for user := range in {
		switch user := user.(type) {
		case User:
			usersBatch[countInBatch] = user
			countInBatch++
			if countInBatch == GetMessagesMaxUsersBatch {
				countInBatch = 0
				wg.Add(1)
				go func(usersBatch []User, out chan interface{}, wg *sync.WaitGroup) {
					defer wg.Done()
					MsgIDs, err := GetMessages(usersBatch...)
					if err != nil {
						fmt.Printf("Messange error!\n")
					}
					for i := 0; i < len(MsgIDs); i++ {
						out <- MsgIDs[i]
					}
				}(usersBatch, out, wg)
				usersBatch = make([]User, GetMessagesMaxUsersBatch)
			}
		default:
			fmt.Printf("Not User!\n")
		}
	}
	if countInBatch != 0 {
		MsgIDs, err := GetMessages(usersBatch[:countInBatch]...)
		if err != nil {
			fmt.Printf("Messange error!\n")
		}
		for i := 0; i < len(MsgIDs); i++ {
			out <- MsgIDs[i]
		}
	}
	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	wg := &sync.WaitGroup{}
	for i := 0; i < HasSpamMaxAsyncRequests; i++ {
		wg.Add(1)
		go func(in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			for msgID := range in {
				switch msgID := msgID.(type) {
				case MsgID:
					res, err := HasSpam(msgID)
					out <- MsgData{
						ID:      msgID,
						HasSpam: res,
					}
					if err != nil {
						fmt.Printf("Too many!")
					}
				default:
					fmt.Printf("Not MsgID!")
				}
			}
		}(in, out, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	msgDatas := make([]MsgData, 0)
	for data := range in {
		switch data := data.(type) {
		case MsgData:
			msgDatas = append(msgDatas, data)
		default:
			fmt.Printf("Wrong type!")
		}
	}
	sort.Slice(msgDatas, func(i, j int) bool {
		if msgDatas[i].HasSpam != msgDatas[j].HasSpam {
			return msgDatas[i].HasSpam && !msgDatas[j].HasSpam
		} else {
			return msgDatas[i].ID < msgDatas[j].ID
		}
	})

	for _, val := range msgDatas {
		out <- strconv.FormatBool(val.HasSpam) + " " + strconv.FormatUint(uint64(val.ID), 10)
	}
}
