package main

import (
	"fmt"
	"github.com/kqbi/gossip/log"
)

var (
	// Caller parameters
	caller = &endpoint{
		displayName: "Ryan",
		username:    "ryan",
		host:        "127.0.0.1",
		port:        5070,
		proxy:       "10.26.1.41",
		proxy_port:  6000,
		transport:   "TCP",
	}

	// Callee parameters
	callee = &endpoint{
		displayName: "Ryan's PC",
		username:    "stefan",
		host:        "127.0.0.1",
		port:        5080,
		proxy:       "10.26.1.21",
		proxy_port:  6010,
		transport:   "TCP",
	}
)

func main() {
	var err error
	log.SetDefaultLogLevel(log.SEVERE)
	if err != nil {
		panic(err)
	}

	err = caller.Start()
	if err != nil {
		panic(err)
	}
	/*
			err = callee.Start()
			if err != nil {
				panic(err)
			}

			f, err := os.Create("cpuprof.out")
			if err != nil {
				panic(err)
			}
			runtime.SetBlockProfileRate(0)
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()

			sub_num := 91000
			caller.dialog.cseq = 1
			successes := 0
			tries := 0
			thissecond := 0
			persecond := 10000
			start := time.Now()
			success := make(chan struct{}, 10)
			go func() {
				for _ = range success {
					successes++
				}
			}()
			end := time.After(20 * time.Second)
			second := time.NewTicker(1 * time.Second)
			go callee.ServeNonInvite()
			go func() {
				for {
					tries++
					thissecond++
					sub_num++
					caller.dialog.callId = fmt.Sprintf("callid%v", sub_num)
					caller.username = fmt.Sprintf("user%v", sub_num)
					caller.dialog.cseq = uint32(sub_num)
					go func() {
						err := caller.Register(callee)
						if err == nil {
							fmt.Printf("Registration success\n")
							success <- struct{}{}
						} else {
							//fmt.Printf("Registration failed: %v\n", err.Error())
						}
					}()

					if thissecond == persecond {
						thissecond = 0
						<-second.C
						fmt.Printf("%v tries. %v successes.\n", tries, successes)
						fmt.Printf("Currently %v goroutines exist.\n", runtime.NumGoroutine())
						fmt.Printf("Time since start: %v\n", time.Since(start))
					}
				}
			}()
		loop:
			for {
				select {
				case <-end:
					f, _ := os.Create("memprof.out")
					pprof.WriteHeapProfile(f)
					f.Close()
					f, _ = os.Create("blockprof.out")
					pprof.Lookup("block").WriteTo(f, 0)
					f.Close()
					break loop
				default:
				}
			}
	*/
	caller.dialog.callId = fmt.Sprintf("callid%v", 1)
	caller.username = fmt.Sprintf("user%v", 1)
	caller.dialog.cseq = uint32(1)
	err = caller.Register(callee)
	if err == nil {
		fmt.Printf("Registration success\n")
	} else {
		//fmt.Printf("Registration failed: %v\n", err.Error())
	}

	select {}
}