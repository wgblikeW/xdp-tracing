package main

/*
#include<load-bpf.h>
*/
import "C"

func main() {

	C.execute_bpf_prog()
}
