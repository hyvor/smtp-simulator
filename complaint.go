package main

import "time"

func sendComplaint(originalMailFrom string, to string, delay int) {

	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}

}
