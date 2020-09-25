# speedtracker
Collect basic analytics on your network performance over time.

This is an extremely simple CLI tool to run in the background.

## Requires

Homebrew - https://brew.sh/

## Install

Run:

```bash
$ brew tap jchengjr77/homebrew-private https://github.com/jchengjr77/homebrew-private.git
$ brew install speedtracker
```

## Usage

For default behavior (ping every 15 seconds):
```
$ speedtracker
```

#### NOTE: It takes around 10-30s to start up. Please be patient while starting...


To modify ping intervals (seconds): 
```
$ speedtracker --interval 60
```
Personally, I run this in the background so I use a `600` second delay.

For help:
```
$ speedtracker -h
```
