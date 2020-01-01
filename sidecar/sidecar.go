package sidecar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// SideCar :nodoc:
type SideCar struct {
	balancerURL string
}

// NewSideCar :nodoc:
func NewSideCar(balancerURL string) *SideCar {
	return &SideCar{balancerURL}
}

// Join joind load balancer cluster
func (sc *SideCar) Join(ports ...string) error {
	if len(ports) == 0 {
		url := sc.balancerURL + "/rebalance/join?port=80"
		return join(url)
	}

	for _, port := range ports {
		url := fmt.Sprintf("%s%s", sc.balancerURL+"/rebalance/join?port=", port)
		if err := join(url); err != nil {
			return err
		}
	}

	return nil
}

func join(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("faild to join")
	}

	log.Println(string(body))

	return nil
}
