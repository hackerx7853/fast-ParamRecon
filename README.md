# fast-ParamRecon

Search URLs for parameters to identify specific vulnerabilities "RCE, XSS, SQLI, REDIRECT, LFI, SSRF"
Can scan 10 million URLs in under 2 minutes "Based on the OWASP Top 25 Parameters"

# The idea behind this is to have many URLs with potentially vulnerable parameters and complement external tools or your own scripts.
 :spades:
 

Install

```bash
go install github.com/hackerx7853/fast-ParamRecon@latest```

```bash curl -O https://raw.githubusercontent.com/hackerx7853/fast-ParamRecon/main/params.json```


Scan: fast-ParamRecon -urls your_urls.txt -params params.json
