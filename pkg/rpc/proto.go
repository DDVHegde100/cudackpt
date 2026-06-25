package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	OpPing      uint32 = 1
	OpStatus    uint32 = 2
	OpFreeze    uint32 = 3
	OpSnapshot  uint32 = 4
	OpRestore   uint32 = 5
	OpResume    uint32 = 6
	OpQuit      uint32 = 7
)

type Client struct {
	conn net.Conn
}

func Dial(pid int) (*Client, error) {
	path := fmt.Sprintf("/run/cudackpt/%d.sock", pid)
	c, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return &Client{conn: c}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func writeU32(w io.Writer, v uint32) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	_, err := w.Write(b[:])
	return err
}

func readU32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

func writeStr(w io.Writer, s string) error {
	if err := writeU32(w, uint32(len(s))); err != nil {
		return err
	}
	if len(s) == 0 {
		return nil
	}
	_, err := w.Write([]byte(s))
	return err
}

func (c *Client) call(op uint32, path string) (uint32, error) {
	if err := writeU32(c.conn, op); err != nil {
		return 1, err
	}
	if op == OpSnapshot || op == OpRestore {
		if err := writeStr(c.conn, path); err != nil {
			return 1, err
		}
	}
	return readU32(c.conn)
}

func (c *Client) Ping() error {
	rc, err := c.call(OpPing, "")
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("ping rc=%d", rc)
	}
	return nil
}

func (c *Client) Status() (uint32, error) {
	return c.call(OpStatus, "")
}

func (c *Client) Freeze() error {
	rc, err := c.call(OpFreeze, "")
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("freeze rc=%d", rc)
	}
	return nil
}

func (c *Client) Snapshot(dir string) error {
	rc, err := c.call(OpSnapshot, dir)
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("snapshot rc=%d", rc)
	}
	return nil
}

func (c *Client) Restore(dir string) error {
	rc, err := c.call(OpRestore, dir)
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("restore rc=%d", rc)
	}
	return nil
}

func (c *Client) Resume() error {
	rc, err := c.call(OpResume, "")
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("resume rc=%d", rc)
	}
	return nil
}
