# fessboxconfig

Is a configuration manager for fessbox

# Developing

You need a working Golang environment

First install
```bash
go get github.com/FarmRadioHangar/fessboxconfig/...
```


Charge to the root of the installed repository

```bash
cd $GOPATH/github.com/FarmRadioHangar/fessboxconfig
```

If you have asterisk installed and want to run in production mode

```bash
make start
```

If you just want to test in dev mode( No asterisk installation is required)

```bash
make dev
```
