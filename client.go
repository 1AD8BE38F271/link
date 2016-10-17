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

type Client struct {
	manager      *Manager

	protocol     Protocol
	dialer       Dialer

	MinSess      uint64

	//data (sents+received) per secord
	MaxSpeed     uint64

	sendChanSize int

	sessions     chan *Session
	isClosed     chan interface{}
}

func NewClient(d Dialer, p Protocol, MinSess, MaxSpeed uint64, sendChanSize int) *Client {
	if MinSess == 0 {
		MinSess = 1
	}
	if MaxSpeed == 0 {
		MaxSpeed = 15
	}

	cli := &Client{
		manager : NewManager(),
		protocol:p,
		dialer:d,
		MinSess: MinSess,
		MaxSpeed: MaxSpeed,
		sendChanSize:sendChanSize,
		sessions:make(chan *Session, 10),
		isClosed:make(chan interface{}),
	}

	return cli
}

func (cli *Client) Serve(handler Handler) error {
	for {
		select {
		case session := <-cli.sessions:
			go handler.HandleSession(session)
		case <-cli.isClosed:
			return nil
		}
	}
}

func (cli *Client) Stop() {
	close(cli.isClosed)
	cli.manager.Dispose()
}

func (cli *Client) createSession() (*Session, error) {
	codec, err := CreateCodec(cli.dialer, cli.protocol)
	if err != nil {
		return nil, err
	}

	s := cli.manager.NewSession(codec, cli.sendChanSize)
	cli.sessions <- s
	return s, nil
}

func (cli *Client) GetSession() (sess *Session, err error) {
	if cli.manager.GetSize() == 0 {
		sess, err = cli.createSession()
		if err != nil {
			return nil, err
		} else {
			return sess, nil
		}
	}

	sess, speed := cli.getLessBusy()
	if sess == nil {
		return nil, ErrNoSession
	} else {
		if speed > cli.MaxSpeed || uint64(cli.manager.GetSize()) < cli.MinSess {
			go cli.createSession()
		}
		return sess, nil
	}

}

func (cli *Client) getLessBusy() (sess *Session, speed uint64) {
	speed = 0
	for _, s := range cli.manager.GetSessions() {
		if speed == 0 || s.GetSpeed() < speed {
			sess = s
			speed = s.GetSpeed()
		}
	}
	return
}
