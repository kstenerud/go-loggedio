// loggedio proxies calls to io.Reader, io.Writer, io.Closer, and net.Conn
// interfaces, reporting their read, write, error, and close events.
//
// The proxied object is not checked for compatibility. If you attempt to call
// a proxied method that the object doesn't actually implement, it will panic.
// It's recommended to cast to the expected interface before use.
//
// Loggedio supports reporting to writers and the go log out of the box. Other
// reporting mechanisms can easily be added using `loggedio.Generic()`.
package loggedio

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// Generic creates a new logged I/O proxy where all reporting behavior is
// user-defined via callback functions.
func Generic(proxiedObject interface{},
	reportReadEvent, reportWriteEvent func(b []byte),
	reportErrorEvent func(location string, err error),
	reportCloseEvent func()) *LoggedIOProxy {
	this := new(LoggedIOProxy)
	this.proxiedObject = proxiedObject
	this.reportReadEvent = reportReadEvent
	this.reportWriteEvent = reportWriteEvent
	this.reportErrorEvent = reportErrorEvent
	this.reportCloseEvent = reportCloseEvent
	return this
}

// StringToLog creates a logged I/O proxy that writes the contents of the data
// as strings to the go log. readFmt and writeFmt must contain a single %v for
// the payload contents. errFmt must contain a %v for the location where the
// error occured, and a second %v for the error payload (in that order).
func StringToLog(proxiedObject interface{},
	readFmt, writeFmt, errorFmt, closeMsg string) *LoggedIOProxy {
	return Generic(proxiedObject,
		func(b []byte) { log.Printf(readFmt, string(b)) },
		func(b []byte) { log.Printf(writeFmt, string(b)) },
		func(location string, err error) { log.Printf(errorFmt, location, err) },
		func() { log.Printf("%v", closeMsg) })
}

// HexToLog creates a logged I/O proxy that writes the hex encoded contents of
// the data to the go log. readFmt and writeFmt must contain a single %v for the
// payload contents. errFmt must contain a %v for the location where the error
// occured, and a second %v for the error payload, in that order.
func HexToLog(proxiedObject interface{},
	readFmt, writeFmt, errorFmt, closeMsg string) *LoggedIOProxy {
	return Generic(proxiedObject,
		func(b []byte) { log.Printf(readFmt, toHex(b)) },
		func(b []byte) { log.Printf(writeFmt, toHex(b)) },
		func(location string, err error) { log.Printf(errorFmt, location, err) },
		func() { log.Printf("%v", closeMsg) })
}

// StringToWriter creates a logged I/O proxy that writes the contents of the
// data as strings to the specified writer. readFmt and writeFmt must contain a
// single %v for the payload contents. errFmt must contain a %v for the location
// where the error occured, and a second %v for the error payload, in that order.
func StringToWriter(proxiedObject interface{}, writer io.Writer,
	readFmt, writeFmt, errorFmt, closeMsg string) *LoggedIOProxy {
	return Generic(proxiedObject,
		func(b []byte) { fmt.Fprintf(writer, readFmt, string(b)) },
		func(b []byte) { fmt.Fprintf(writer, writeFmt, string(b)) },
		func(location string, err error) { fmt.Fprintf(writer, errorFmt, location, err) },
		func() { fmt.Fprintf(writer, "%v", closeMsg) })
}

// HexToWriter creates a logged I/O proxy that writes the hex encoded contents
// of the data to the specified writer. readFmt and writeFmt must contain a
// single %v for the payload contents. errFmt must contain a %v for the location
// where the error occured, and a second %v for the error payload, in that order.
func HexToWriter(proxiedObject interface{}, writer io.Writer,
	readFmt, writeFmt, errorFmt, closeMsg string) *LoggedIOProxy {
	return Generic(proxiedObject,
		func(b []byte) { fmt.Fprintf(writer, readFmt, toHex(b)) },
		func(b []byte) { fmt.Fprintf(writer, writeFmt, toHex(b)) },
		func(location string, err error) { fmt.Fprintf(writer, errorFmt, location, err) },
		func() { fmt.Fprintf(writer, "%v", closeMsg) })
}

// LoggedIOProxy implements io.Reader, io.Writer, io.Closer, and net.Conn,
// proxying their API and calling back on read, write, error, and close events.
// Callbacks are called AFTER the event occurs. If an error occurs on a read or
// write, only the bytes actually read/written will be reported (if > 0), after
// which the error will be reported.
type LoggedIOProxy struct {
	reportReadEvent  func(readContents []byte)
	reportWriteEvent func(writeContents []byte)
	reportCloseEvent func()
	reportErrorEvent func(location string, err error)
	proxiedObject    interface{}
	location         string
}

func (this *LoggedIOProxy) Read(b []byte) (n int, err error) {
	reader := this.proxiedObject.(io.Reader)
	n, err = reader.Read(b)
	if n > 0 {
		this.reportReadEvent(b[:n])
	}
	if err != nil {
		this.reportErrorEvent("Read()", err)
	}
	return
}

func (this *LoggedIOProxy) Write(b []byte) (n int, err error) {
	writer := this.proxiedObject.(io.Writer)
	n, err = writer.Write(b)
	if n > 0 {
		this.reportWriteEvent(b[:n])
	}
	if err != nil {
		this.reportErrorEvent("Write()", err)
	}
	return
}

func (this *LoggedIOProxy) Close() (err error) {
	closer := this.proxiedObject.(io.Closer)
	err = closer.Close()
	this.reportCloseEvent()
	if err != nil {
		this.reportErrorEvent("Close()", err)
	}
	return
}

func (this *LoggedIOProxy) LocalAddr() net.Addr {
	conn := this.proxiedObject.(net.Conn)
	return conn.LocalAddr()
}

func (this *LoggedIOProxy) RemoteAddr() net.Addr {
	conn := this.proxiedObject.(net.Conn)
	return conn.RemoteAddr()
}

func (this *LoggedIOProxy) SetDeadline(t time.Time) (err error) {
	conn := this.proxiedObject.(net.Conn)
	err = conn.SetDeadline(t)
	if err != nil {
		this.reportErrorEvent("SetDeadline()", err)
	}
	return
}

func (this *LoggedIOProxy) SetReadDeadline(t time.Time) (err error) {
	conn := this.proxiedObject.(net.Conn)
	err = conn.SetReadDeadline(t)
	if err != nil {
		this.reportErrorEvent("SetReadDeadline()", err)
	}
	return
}

func (this *LoggedIOProxy) SetWriteDeadline(t time.Time) (err error) {
	conn := this.proxiedObject.(net.Conn)
	err = conn.SetWriteDeadline(t)
	if err != nil {
		this.reportErrorEvent("SetWriteDeadline()", err)
	}
	return
}

var hexDigits = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f',
}

func toHex(b []byte) string {
	builder := strings.Builder{}
	for i := 0; i < len(b); i++ {
		ch := b[i]
		builder.WriteByte(hexDigits[ch>>4])
		builder.WriteByte(hexDigits[ch&15])
		if i < len(b)-1 {
			builder.WriteByte(' ')
		}
	}
	return builder.String()
}
