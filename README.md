Logged I/O for Go
=================

A go module to proxy I/O operations so that they can be reported.

loggedio proxies calls to io.Reader, io.Writer, io.Closer, and net.Conn
interfaces, reporting their read, write, error, and close events.

The proxied object is not checked for compatibility. If you attempt to call
a proxied method that the object doesn't actually implement, it will panic.
It's recommended to cast to the expected interface before use.

Loggedio supports reporting to writers and the go log out of the box. Other
reporting mechanisms can easily be added using `loggedio.Generic()`.


Usage
-----

```golang
func Demonstrate() {
	buffer := make([]byte, 8, 8)
	writer := &bytes.Buffer{}
	proxy := loggedio.StringToWriter(writer, os.Stdout, "R [%v]\n", "W [%v]\n", "E [%v: %v]\n", "C\n")

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
	proxy = loggedio.HexToWriter(writer, os.Stdout, "R [%v]\n", "W [%v]\n", "E [%v: %v]\n", "C\n")
	proxy.Write([]byte{1, 2, 0xae, 0xf1})
}
```

Output:

```
W [Testing!]
R [Testing!]
W [01 02 ae f1]
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
