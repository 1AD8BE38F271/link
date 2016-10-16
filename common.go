/*
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * Author: FTwOoO <booobooob@gmail.com>
 */

package link

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrNoSession = errors.New("session in pool but can't pick one.")
	ErrSessionNotFound = errors.New("session not found.")
)

var (
	UPDATE_INTERVAL time.Duration = 10
)

type SpeedCounter struct {
	cnt      uint32
	Speed    uint32
	All      uint64
	isClosed chan interface{}
}

func NewSpeedCounter() (sc *SpeedCounter) {
	sc = &SpeedCounter{isClosed:make(chan interface{}, 1)}
	go sc.Update()
	return sc
}

func (sc *SpeedCounter) Close() error {
	sc.isClosed <- 1
	return nil
}

func (sc *SpeedCounter) Update() {
	hsHeartbeat := time.Tick(UPDATE_INTERVAL * time.Second)
	for {
		select {
		case <-hsHeartbeat:
			c := atomic.SwapUint32(&sc.cnt, 0)
			sc.Speed = c / uint32(UPDATE_INTERVAL)
			sc.All += uint64(c)
		case <-sc.isClosed:
			return
		}
	}
}

func (sc *SpeedCounter) Add(s uint32) uint32 {
	return atomic.AddUint32(&sc.cnt, s)
}