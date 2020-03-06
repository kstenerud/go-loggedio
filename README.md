Logged I/O for Go
=================

A go module to proxy I/O operations so that they can be reported.

LoggedIO proxies calls to `io.Reader`, `io.Writer`, `io.Closer`, and `net.Conn`
interfaces, reporting their read, write, error, and close events.

LoggedIO uses duck typing, meaning that the proxied object is not checked
for compatibility until you actually call a method. If you attempt to call
a proxied method that the object doesn't actually implement, it will panic.
It's recommended to cast to the expected interface before use for better type
safety.

The following proxy generators are available:

* **Generic:** All reporting behavior is provided by user-defined functions.
* **StringToLog:** Interprets all data as strings and writes them to the go log.
* **HexToLog:** Converts all data to hex and writes them to the go log.
* **StringToLog:** Interprets all data as strings and writes them to the specified `io.Writer`.
* **HexToLog:** Converts all data to hex and writes them to the specified `io.Writer`.
* **DumpToWriters:** Dumps all reads and writes to separate `io.Writer` objects.
* **DumpToFiles:** Dumps all reads and writes to separate files.


Usage
-----

```golang
func Demonstrate() {
	readerWriter := &bytes.Buffer{}
	var err error

	// Log all reads/writes as hex to stdout:
	proxy := loggedio.HexToWriter(readerWriter, os.Stdout, "R [%v]\n", "W [%v]\n", "E [%v: %v]\n", "C\n")
	if _, err = proxy.Write([]byte{1, 2, 0xae, 0xf1}); err != nil {
		// TODO
	}

	readerWriter.Reset()

	// Log all reads/writes as both hex AND strings (chaining together two proxies):
	stringProxy := loggedio.StringToWriter(readerWriter, os.Stdout, "RS [%v]\n", "WS [%v]\n", "E [%v: %v]\n", "C\n")
	hexProxy := loggedio.HexToWriter(stringProxy, os.Stdout, "RH [%v]\n", "WH [%v]\n", "", "")

	if _, err = hexProxy.Write([]byte("Testing")); err != nil {
		// TODO
	}

	if _, err = hexProxy.Write([]byte(" 1 2 3!")); err != nil {
		// TODO
	}

	buffer := make([]byte, readerWriter.Len())
	if _, err = hexProxy.Read(buffer); err != nil {
		// TODO
	}
}
```

Output:

```
W [01 02 ae f1]
WS [Testing]
WH [54 65 73 74 69 6e 67]
WS [ 1 2 3!]
WH [20 31 20 32 20 33 21]
RS [Testing 1 2 3!]
RH [54 65 73 74 69 6e 67 20 31 20 32 20 33 21]
```

For other examples, see the [unit tests](loggedio_test.go).


License
-------

MIT License:

Copyright 2020 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
