// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"errors"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

type clientSession struct {
	session
}

func (s *clientSession) Connect() error {
	return errors.New("not implemented")
}
func (s *clientSession) HandleConnack(pkt *packet.ConnackPacket) error {
	return errors.New("not implemented")
}
func (s *clientSession) Subscribe(map[string]byte) error {
	return errors.New("not implemented")
}
func (s *clientSession) HandleSuback(pkt *packet.SubackPacket) error {
	return errors.New("not implemented")
}
func (s *clientSession) Unsubscribe(...string) error {
	return errors.New("not implemented")
}
func (s *clientSession) HandleUnsuback(pkt *packet.UnsubackPacket) error {
	return errors.New("not implemented")
}
