# HG8045Q

Show devices connected to HG8045Q

## Installation

```console
$ go install github.com/dytlzl/hg8045q@latest
```

## Usage

```console
$ HG8045Q_USERNAME=admin HG8045Q_PASSWORD=password go run .
GLOBAL IP: 39.117.xx.xx
IP               MAC ADDRESS         STATUS    HOSTNAME
192.168.1.11     00:1c:fc:27:ee:eb   Online    iPhone  
192.168.1.12     a0:83:e0:60:22:a0   Online    MacBook 
192.168.1.13     a0:66:70:32:49:b0   Offline   iMac    
192.168.1.14     40:60:10:32:7d:90   Offline   iPad    
```
