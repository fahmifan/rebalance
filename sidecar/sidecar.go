package sidecar

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/miun173/rebalance/sidecar/config"
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
func (sc *SideCar) Join() error {
	resp, err := http.Get(sc.balancerURL + "/rebalance/join?port=" + config.ClientServicePort())
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
