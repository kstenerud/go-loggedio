package loggedio

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func reportPanic(function func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("%v", e)
			}
		}
	}()

	function()
	return
}

func assertNoPanic(t *testing.T, function func()) {
	if err := reportPanic(function); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func assertPanics(t *testing.T, function func()) {
	if err := reportPanic(function); err == nil {
		t.Errorf("Expected an error")
	}
}

func generateBytes(length int) []byte {
	result := make([]byte, length, length)
	charRange := int('z' - 'a')
	for i := 0; i < length; i++ {
		result[i] = byte(i%charRange + 'a')
	}
	return result
}

func generateError() error {
	return fmt.Errorf("ERROR!")
}

type NullWriter struct{}

func (this *NullWriter) Write(b []byte) (n int, err error) { return }

type MockIO struct {
	WriteContents             []byte
	CloseCallCount            int
	LocalAddrCallCount        int
	RemoteAddrCallCount       int
	SetDeadlineCallCount      int
	SetReadDeadlineCallCount  int
	SetWriteDeadlineCallCount int
	FailAfterWriteByteCount   int
	FailAfterReadByteCount    int
	FailNextOperations        bool
}

func (this *MockIO) Read(b []byte) (n int, err error) {
	if this.FailNextOperations {
		err = generateError()
		return
	}

	n = len(b)
	if n >= this.FailAfterReadByteCount && this.FailAfterReadByteCount > 0 {
		n = this.FailAfterReadByteCount
	}
	copy(b, generateBytes(n))
	if n >= this.FailAfterReadByteCount && this.FailAfterReadByteCount > 0 {
		err = generateError()
	}
	return
}

func (this *MockIO) Write(b []byte) (n int, err error) {
	if this.FailNextOperations {
		err = generateError()
		return
	}

	n = len(b)
	if n >= this.FailAfterWriteByteCount && this.FailAfterWriteByteCount > 0 {
		n = this.FailAfterWriteByteCount
	}

	b = b[:n]
	this.WriteContents = append(this.WriteContents, b...)
	if n >= this.FailAfterWriteByteCount && this.FailAfterWriteByteCount > 0 {
		err = generateError()
	}

	return
}

func (this *MockIO) Close() (err error) {
	this.CloseCallCount++
	if this.FailNextOperations {
		err = generateError()
		return
	}

	return
}

func (this *MockIO) LocalAddr() net.Addr {
	this.LocalAddrCallCount++
	return nil
}

func (this *MockIO) RemoteAddr() net.Addr {
	this.RemoteAddrCallCount++
	return nil
}

func (this *MockIO) SetDeadline(t time.Time) (err error) {
	this.SetDeadlineCallCount++
	if this.FailNextOperations {
		err = generateError()
		return
	}
	return
}

func (this *MockIO) SetReadDeadline(t time.Time) (err error) {
	this.SetReadDeadlineCallCount++
	if this.FailNextOperations {
		err = generateError()
		return
	}
	return
}

func (this *MockIO) SetWriteDeadline(t time.Time) (err error) {
	this.SetWriteDeadlineCallCount++
	if this.FailNextOperations {
		err = generateError()
		return
	}
	return
}

type MockReader struct {
	implementation *MockIO
}

func (this *MockReader) Read(b []byte) (n int, err error) {
	return this.implementation.Read(b)
}

type MockWriter struct {
	implementation *MockIO
}

func (this *MockWriter) Write(b []byte) (n int, err error) {
	return this.implementation.Write(b)
}

type MockCloser struct {
	implementation *MockIO
}

func (this *MockCloser) Close() (err error) {
	return this.implementation.Close()
}

// -----------------------------------------------------------------------------

func expectNumber(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func expectLength(t *testing.T, data []byte, length int) {
	if length != len(data) {
		t.Errorf("Expected data to be length %v but got %v", len(data), length)
	}
}

func expectBufferContents(t *testing.T, buffer *bytes.Buffer, expected string) {
	if buffer.String() != expected {
		t.Errorf("Expected buffer to contain \"%v\" but found \"%v\"", expected, buffer.String())
		return
	}
}

func expectError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error")
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestReadWriteString(t *testing.T) {
	proxied := &MockIO{}
	buffer := &bytes.Buffer{}
	logged := StringToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")

	readBuffer := generateBytes(3)
	n, err := logged.Read(readBuffer)
	expectNoError(t, err)
	expectLength(t, []byte(readBuffer), n)
	expectBufferContents(t, buffer, "R [abc]")

	buffer.Reset()
	writeValue := "test"
	n, err = logged.Write([]byte(writeValue))
	expectNoError(t, err)
	expectLength(t, []byte(writeValue), n)
	expectBufferContents(t, buffer, "W [test]")

	logged = StringToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	n, err = logged.Read(readBuffer)
	expectNoError(t, err)
	expectLength(t, []byte(readBuffer), n)
	n, err = logged.Write([]byte(writeValue))
	expectNoError(t, err)
	expectLength(t, []byte(writeValue), n)
}

func TestReadWriteHex(t *testing.T) {
	proxied := &MockIO{}
	buffer := &bytes.Buffer{}
	logged := HexToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")

	readBuffer := generateBytes(3)
	n, err := logged.Read(readBuffer)
	expectNoError(t, err)
	expectLength(t, []byte(readBuffer), n)
	expectBufferContents(t, buffer, "R [61 62 63]")

	buffer.Reset()
	writeValue := []byte{1, 2, 3}
	n, err = logged.Write([]byte(writeValue))
	expectNoError(t, err)
	expectLength(t, []byte(writeValue), n)
	expectBufferContents(t, buffer, "W [01 02 03]")

	logged = HexToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	n, err = logged.Read(readBuffer)
	expectNoError(t, err)
	expectLength(t, []byte(readBuffer), n)
	n, err = logged.Write([]byte(writeValue))
	expectNoError(t, err)
	expectLength(t, []byte(writeValue), n)
}

func TestClose(t *testing.T) {
	proxied := &MockIO{}
	buffer := &bytes.Buffer{}

	logged := HexToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	expectNoError(t, logged.Close())
	expectBufferContents(t, buffer, "C")
	expectNumber(t, 1, proxied.CloseCallCount)

	proxied = &MockIO{}
	buffer.Reset()
	logged = StringToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	expectNoError(t, logged.Close())
	expectBufferContents(t, buffer, "C")
	expectNumber(t, 1, proxied.CloseCallCount)

	proxied = &MockIO{}
	logged = StringToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	expectNoError(t, logged.Close())
	expectNumber(t, 1, proxied.CloseCallCount)

	proxied = &MockIO{}
	logged = HexToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	expectNoError(t, logged.Close())
	expectNumber(t, 1, proxied.CloseCallCount)
}

func TestOtherOps(t *testing.T) {
	proxied := &MockIO{}
	buffer := &bytes.Buffer{}

	logged := StringToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	logged.LocalAddr()
	expectNumber(t, 1, proxied.LocalAddrCallCount)
	logged.RemoteAddr()
	expectNumber(t, 1, proxied.RemoteAddrCallCount)
	expectNoError(t, logged.SetDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetDeadlineCallCount)
	expectNoError(t, logged.SetReadDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetReadDeadlineCallCount)
	expectNoError(t, logged.SetWriteDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetWriteDeadlineCallCount)

	proxied = &MockIO{}
	logged = HexToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	logged.LocalAddr()
	expectNumber(t, 1, proxied.LocalAddrCallCount)
	logged.RemoteAddr()
	expectNumber(t, 1, proxied.RemoteAddrCallCount)
	expectNoError(t, logged.SetDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetDeadlineCallCount)
	expectNoError(t, logged.SetReadDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetReadDeadlineCallCount)
	expectNoError(t, logged.SetWriteDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetWriteDeadlineCallCount)

	proxied = &MockIO{}
	logged = StringToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	logged.LocalAddr()
	expectNumber(t, 1, proxied.LocalAddrCallCount)
	logged.RemoteAddr()
	expectNumber(t, 1, proxied.RemoteAddrCallCount)
	expectNoError(t, logged.SetDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetDeadlineCallCount)
	expectNoError(t, logged.SetReadDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetReadDeadlineCallCount)
	expectNoError(t, logged.SetWriteDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetWriteDeadlineCallCount)

	proxied = &MockIO{}
	logged = HexToLog(proxied, "R [%v]", "W [%v]", "E [%v: %v]", "C")
	logged.LocalAddr()
	expectNumber(t, 1, proxied.LocalAddrCallCount)
	logged.RemoteAddr()
	expectNumber(t, 1, proxied.RemoteAddrCallCount)
	expectNoError(t, logged.SetDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetDeadlineCallCount)
	expectNoError(t, logged.SetReadDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetReadDeadlineCallCount)
	expectNoError(t, logged.SetWriteDeadline(time.Now()))
	expectNumber(t, 1, proxied.SetWriteDeadlineCallCount)
}

func testFail(t *testing.T, buffer *bytes.Buffer, logged *LoggedIOProxy) {
	buffer.Reset()
	n, err := logged.Read([]byte{1})
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "E [Read(): ERROR!]")

	buffer.Reset()
	n, err = logged.Write([]byte{1})
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "E [Write(): ERROR!]")

	buffer.Reset()
	err = logged.Close()
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "CE [Close(): ERROR!]")

	buffer.Reset()
	err = logged.SetDeadline(time.Now())
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "E [SetDeadline(): ERROR!]")

	buffer.Reset()
	err = logged.SetReadDeadline(time.Now())
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "E [SetReadDeadline(): ERROR!]")

	buffer.Reset()
	err = logged.SetWriteDeadline(time.Now())
	expectError(t, err)
	expectNumber(t, 0, n)
	expectBufferContents(t, buffer, "E [SetWriteDeadline(): ERROR!]")
}

func TestFail(t *testing.T) {
	proxied := &MockIO{FailNextOperations: true}
	buffer := &bytes.Buffer{}

	testFail(t, buffer, StringToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C"))
	testFail(t, buffer, HexToWriter(proxied, buffer, "R [%v]", "W [%v]", "E [%v: %v]", "C"))
}

func TestWrongInterface(t *testing.T) {
	var intf interface{}
	var logged *LoggedIOProxy
	nullWriter := &NullWriter{}

	intf = &MockReader{implementation: &MockIO{}}
	logged = StringToWriter(intf, nullWriter, "%v", "%v", "%v", "")
	assertNoPanic(t, func() { logged.Read([]byte{0}) })
	assertPanics(t, func() { logged.Write([]byte{1}) })
	assertPanics(t, func() { logged.LocalAddr() })
	assertPanics(t, func() { logged.RemoteAddr() })
	assertPanics(t, func() { logged.SetDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetReadDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetWriteDeadline(time.Now()) })
	assertPanics(t, func() { logged.Close() })

	intf = &MockWriter{implementation: &MockIO{}}
	logged = StringToWriter(intf, nullWriter, "%v", "%v", "%v", "")
	assertPanics(t, func() { logged.Read([]byte{0}) })
	assertNoPanic(t, func() { logged.Write([]byte{1}) })
	assertPanics(t, func() { logged.LocalAddr() })
	assertPanics(t, func() { logged.RemoteAddr() })
	assertPanics(t, func() { logged.SetDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetReadDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetWriteDeadline(time.Now()) })
	assertPanics(t, func() { logged.Close() })

	intf = &MockCloser{implementation: &MockIO{}}
	logged = StringToWriter(intf, nullWriter, "%v", "%v", "%v", "")
	assertPanics(t, func() { logged.Read([]byte{0}) })
	assertPanics(t, func() { logged.Write([]byte{1}) })
	assertPanics(t, func() { logged.LocalAddr() })
	assertPanics(t, func() { logged.RemoteAddr() })
	assertPanics(t, func() { logged.SetDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetReadDeadline(time.Now()) })
	assertPanics(t, func() { logged.SetWriteDeadline(time.Now()) })
	assertNoPanic(t, func() { logged.Close() })
}

func Demonstrate() {
	buffer := make([]byte, 8, 8)
	writer := &bytes.Buffer{}
	proxy := StringToWriter(writer, os.Stdout, "R [%v]\n", "W [%v]\n", "E [%v: %v]\n", "C\n")

	n, err := proxy.Write([]byte("Testing!"))
	if err != nil {
		// TODO
	}
	_ = n

	n, err = proxy.Read(buffer)
	if err != nil {
		// TODO
	}
	_ = n

	writer.Reset()
	proxy = HexToWriter(writer, os.Stdout, "R [%v]\n", "W [%v]\n", "E [%v: %v]\n", "C\n")
	proxy.Write([]byte{1, 2, 0xae, 0xf1})
}

func TestDemonstrate(t *testing.T) {
	Demonstrate()
}
