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

type codecCreateFunc func() (Codec, error)

type SessionPool struct {
	Manager

	MinSess      uint64

	//data (sents+received) per secord
	MaxSpeed     uint64

	codecFunc    codecCreateFunc
	sendChanSize int
}

func NewSessionPool(MinSess, MaxSpeed uint64, codecFunc codecCreateFunc, sendChanSize int) (sp *SessionPool) {
	if MinSess == 0 {
		MinSess = 1
	}
	if MaxSpeed == 0 {
		MaxSpeed = 15
	}
	sp = &SessionPool{
		MinSess: MinSess,
		MaxSpeed: MaxSpeed,
		codecFunc:codecFunc,
		sendChanSize:sendChanSize,
	}

	sp.Manager = *NewManager()
	return
}

func (sp *SessionPool) createSession() (*Session, error) {
	codec, err := sp.codecFunc()
	if err != nil {
		return nil, err
	}

	s := sp.NewSession(codec, sp.sendChanSize)
	return s, nil
}

func (sp *SessionPool) Get() (sess *Session, err error) {
	if sp.GetSize() == 0 {
		_, err = sp.createSession()
		if err != nil {
			return nil, err
		}
	}

	sess, speed := sp.getLessBusy()
	if sess == nil {
		return nil, ErrNoSession
	}

	if speed > sp.MaxSpeed || uint64(sp.GetSize()) < sp.MinSess {
		go sp.createSession()
	}
	return
}

func (sp *SessionPool) getLessBusy() (sess *Session, speed uint64) {
	speed = 0
	for _, s := range sp.GetSessions() {
		if speed == 0 || s.GetSpeed() < speed {
			sess = s
			speed = s.GetSpeed()
		}
	}
	return
}




