
# **GFUZZ**

<!---![](https://media.giphy.com/media/lIUbFDEyeBpBLDTuTe/giphy.gif)--->

<p align="center">
<img src="https://media.giphy.com/media/lIUbFDEyeBpBLDTuTe/giphy.gif">
</p>

**gfuzz** aims to reproduce wfuzz's functionality and versatility and combine it with the concurrency of go. It can be used to fuzz GET or POST request headers, including UserAgent, Cookies, and Username and Password for basic auth.

Install: 
`go get github.com/braaaax/gfuzz && go build`  
Basic Usage: 
`gfuzz -z file,wordlist1 -z list,-.php http://someip/FUZZ/FUZ2Z`


```
[+] Author: brax (https://github.com/braaaax/gfuzz)

Usage:   gfuzz [options] <url>
Keyword: FUZZ, ..., FUZnZ  wherever you put these keywords gfuzz will replace them with the values of the specified payload.

Options:
-h/--help                     : This help.
-w wordlist                   : Specify a wordlist file (alias for -z file,wordlist).
-z file/range/list,PAYLOAD    : Where PAYLOAD is FILENAME or 1-10 or "-" separated sequence.
--hc/hl/hw/hh N[,N]+          : Hide responses with the specified code, lines, words, or chars.
--sc/sl/sw/sh N[,N]]+         : Show responses with the specified code, lines, words, or chars.
-t N                          : Specify the number of concurrent connections (10 default).
--post                        : Specify POST request method.
--post-form key=FUZZ          : Specify form value eg key=value.
-p IP:PORT                    : Specify proxy.
-b COOKIE                     : Specify cookie.
-ua USERAGENT                 : Specify user agent.
--password PASSWORD           : Specify password for basic web auth.
--username USERNAME           : Specify username.
--no-follow                   : Don't follow HTTP(S) redirections.
--no-color                    : Monotone output. (use for windows
--print-body                  : Print response body to stdout.
-k                            : Strict TLS connections (skip verify=false opposite of curl).
-q                            : No output.
-H                            : Add headers. (e.g. Key:Value)

Examples: gfuzz -w users.txt -w pass.txt --sc 200 http://www.site.com/log.asp?user=FUZZ&pass=FUZ2Z
          gfuzz -z file,default/common.txt -z list,-.php http://somesite.com/FUZZFUZ2Z
          gfuzz -t 32 -w somelist.txt https://someTLSsite.com/FUZZ
          gfuzz --print-body --sc 200 --post-form "name=FUZZ" -z file,somelist.txt http://somesite.com/form
          gfuzz --post -b mycookie -ua normalbrowser --username admin --password FUZZ -z list,admin-password http://somesite.com
```

