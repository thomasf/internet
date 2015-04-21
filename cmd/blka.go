package main

import (
	"log"
	"net"
)

func main() {
	log.Println(net.IPv4Mask(255, 240, 0, 0))
	for i := 0; i < 8*net.IPv4len; i++ {
		log.Println(net.CIDRMask(8*net.IPv4len-i, 8*net.IPv4len))
	}

	for _, value := range []net.IPMask{
		net.IPv4Mask(255, 255, 255, 255),
		net.IPv4Mask(255, 255, 255, 252),
		net.IPv4Mask(255, 255, 255, 248),
		net.IPv4Mask(255, 255, 255, 240),
		net.IPv4Mask(255, 255, 255, 224),
		net.IPv4Mask(255, 255, 255, 192),
		net.IPv4Mask(255, 255, 255, 128),
		net.IPv4Mask(255, 255, 255, 0),
		net.IPv4Mask(255, 255, 254, 0),
		net.IPv4Mask(255, 255, 252, 0),
		net.IPv4Mask(255, 255, 248, 0),
		net.IPv4Mask(255, 255, 240, 0),
		net.IPv4Mask(255, 255, 224, 0),
		net.IPv4Mask(255, 255, 192, 0),
		net.IPv4Mask(255, 255, 128, 0),
		net.IPv4Mask(255, 255, 0, 0),
		net.IPv4Mask(255, 254, 0, 0),
		net.IPv4Mask(255, 252, 0, 0),
		net.IPv4Mask(255, 248, 0, 0),
		net.IPv4Mask(255, 240, 0, 0),
		net.IPv4Mask(255, 224, 0, 0),
		net.IPv4Mask(255, 192, 0, 0),
		net.IPv4Mask(255, 128, 0, 0),
		net.IPv4Mask(255, 0, 0, 0),
		net.IPv4Mask(254, 0, 0, 0),
		net.IPv4Mask(252, 0, 0, 0),
		net.IPv4Mask(248, 0, 0, 0),
		net.IPv4Mask(240, 0, 0, 0),
		net.IPv4Mask(224, 0, 0, 0),
		net.IPv4Mask(192, 0, 0, 0),
		net.IPv4Mask(128, 0, 0, 0),
	} {
		log.Println(value)
	}

}
