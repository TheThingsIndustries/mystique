# Mystique

[![Build Status](https://travis-ci.com/TheThingsIndustries/mystique.svg?token=1QaLXVRDNDzteUYgpS8B&branch=master)](https://travis-ci.com/TheThingsIndustries/mystique) [![GoDoc](https://godoc.org/github.com/TheThingsIndustries/mystique?status.svg)](https://godoc.org/github.com/TheThingsIndustries/mystique)

Mystique is an MQTT server that implements most parts of the MQTT v3.1.1 specification.

## Getting Started

In Docker: 

- `docker pull thethingsindustries/mystique-server`
- `docker run -d -p 1883:1883 thethingsindustries/mystique-server`

Install and run from source:

- `go get github.com/TheThingsIndustries/mystique/cmd/mystique-server`
- `mystique-server --help`

## Documentation

- [Godoc of `mystique-server` command](https://godoc.org/github.com/TheThingsIndustries/mystique/cmd/mystique-server)
- [Godoc of `ttn-mqtt` command](https://godoc.org/github.com/TheThingsIndustries/mystique/cmd/ttn-mqtt)

## Support

- [Github Issues](https://github.com/TheThingsIndustries/mystique/issues)

## MIT License

**Permissions:** ✅ Commercial use, ✅ Modification, ✅ Distribution, ✅ Private use  
**Limitations:** ❌ Liability, ❌ Warranty  
**Conditions:** ℹ License and copyright notice
