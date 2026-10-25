package i2p

import (
	"bufio"
	"time"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const SigType = 7

// a destination keypair file
type Keyfile struct {
	privkey string
	pubkey  string
	fname   string
}

// save to filesystem
func (k *Keyfile) Store() (err error) {
	if len(k.fname) > 0 {
		var f *os.File
		f, err = os.OpenFile(k.fname, os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			err = k.write(f)
			f.Close()
		}
	}
	return
}

// load from filesystem
func (k *Keyfile) Load() (err error) {
	if len(k.fname) > 0 {
		var f *os.File
		f, err = os.Open(k.fname)
		if err == nil {
			err = k.read(f)
			f.Close()
		}
	}
	return
}

func (k *Keyfile) write(w io.Writer) (err error) {
	_, err = fmt.Fprintf(w, "%s\n%s\n", k.privkey, k.pubkey)
	return
}

func (k *Keyfile) read(r io.Reader) (err error) {
	br := bufio.NewReader(r)
	k.privkey, err = br.ReadString(10)
	k.pubkey, err = br.ReadString(10)
	k.privkey = strings.Trim(k.privkey, "\n")
	k.pubkey = strings.Trim(k.pubkey, "\n")
	return
}

func (k *Keyfile) Addr() Addr {
	return I2PAddr(k.pubkey)
}

// ensure keys are created using a control socket
func (k *Keyfile) ensure(nc net.Conn) (err error) {
	if len(k.fname) > 0 {
		_, err = os.Stat(k.fname)
	}
	if os.IsNotExist(err) || len(k.fname) == 0 {
		// Bound both command transmission and key generation response.
		if err = nc.SetDeadline(time.Now().Add(samCommandTimeout)); err != nil { return err }
		defer nc.SetDeadline(time.Time{})
		_, err = fmt.Fprintf(nc, "DEST GENERATE SIGNATURE_TYPE=%d\n", SigType)
		if err != nil { return err }
		line, readErr := readLine(nc, make([]byte, 1))
		if readErr != nil { return readErr }
		attrs, parseErr := verifySAMFields(line, "DEST", "REPLY")
		if parseErr != nil { return parseErr }
		pub, hasPub := attrs["PUB"]
		priv, hasPriv := attrs["PRIV"]
		if !hasPub || !hasPriv { return fmt.Errorf("DEST REPLY missing PUB or PRIV") }
		k.pubkey, k.privkey = pub, priv
		return k.Store()
	}
	// load keys
	err = k.Load()
	return
}

// create new keyfile given filepath
func NewKeyfile(f string) *Keyfile {
	if strings.ToUpper(f) == "TRANSIENT" {
		f = ""
	}
	return &Keyfile{
		fname: f,
	}
}
