# go-audioserver

<!-- ## Contents
- [About](#about)
- [Usage](#usage)
- [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Commands](#commands)
- [Packet structure](#packet-structure) -->

## About
**Go-audioserver** allows root-level programs to play audio on linux-based systems without the need of changing sound server's configurations and permissions. It is based on a simple IPC socket protocol allowing one-way audio playback control for other processes on a shared system. \
In the future funtionality and purpose might be extended to just allow for general control of custom audio devices in a local network.


## Features
Server supports:
- [x] **Audio playing** - by providing file path or encoded audio data
- [x] **Pause/Resume actions**
- [x] **Early playback termination**

## Usage
### Prerequisites
Make sure you have installed the newest version of go and ALSA installed
- On Ubuntu/Debian:
```
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go libasound2-dev
```
- On Fedora:
```
sudo dnf install golang alsa-lib-devel
```
Add GOPATH to env
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
### Installation
Install server with *go install*
```
go install github.com/not0ff/go-audioserver/cmd/go-audioserver@latest
```
*Optional:* Install cli client
```
go install github.com/not0ff/go-audioserver/cmd/audioserver-cli@latest
```
### Commands
Run local server
```
go-audioserver -sock /path/to/sock # Defaults to /tmp/go-audioserver.sock
```
To check cli client options run:
```
audioserver-cli -h
```

## Packet structure
Structure of raw sent packet
| 4 byte | x bytes |
| ----------- | ----------- |
| Length-prefix | Marshalled json message |

Message structure
| Name | Field | Options | Summary |
| ----------- | ----------- | ----------- | ----------- |
| action | int | 0-3 (inorder: play, pause, resume, quit) |  |
| payload | bytes | PlayPayload(play); IdPayload(pause, resume, quit) | Marshalled json payload |

Payload Fields
| PlayPayload | IdPayload | Type | Required | Info |
| -- | -- | -- | -- | -- |
| id| id | int | Yes | Reference id for playback control |
| format | --- | string | Yes | Format of audio (supported: "wav", "mp3") |
| path | --- | string | No if "data"| Path to audio file on system |
| data | --- | bytes | No if "path" | Audio data compressed with gzip| 
| volume | --- | int | No | Positive number increases volume, nagative decreases  |
| loop | --- | bool | No | Loop playback until quit action  received |
