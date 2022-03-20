## Batty

A tool to make other users of a Linux system Go Batty!

Golang + Bad + TTY = Go Batty


Specifically, it will read and write to a specified TTY, randomly introducing typos as the TTY user inputs characters.

### Usage

1. Identify the tty device for your target TTY
```
$ tty
/dev/pts/1
```
2. Run the `batty` command as root against your target TTY
```
$ sudo ./batty /dev/pts/1
```
3. Type away in the target TTY, watch weird things occur

### Configuration

You may specify a different rate for error induction using the `-r` or `-rate` flat. By default typos will be edited in 10% of the time.
