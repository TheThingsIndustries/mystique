// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

// ConnectReturnCode is returned in the Connack
type ConnectReturnCode byte

// Connect return codes
var (
	ConnectAccepted                    ConnectReturnCode = 0x00
	ConnectUnacceptableProtocolVersion ConnectReturnCode = 0x01
	ConnectIdentifierRejected          ConnectReturnCode = 0x02
	ConnectServerUnavailable           ConnectReturnCode = 0x03
	ConnectMalformedUsernameOrPassword ConnectReturnCode = 0x04
	ConnectNotAuthorized               ConnectReturnCode = 0x05
)

func (c ConnectReturnCode) Error() string {
	switch c {
	case ConnectAccepted: // this is not actually an error
		return "Accepted"
	case ConnectUnacceptableProtocolVersion:
		return "Unacceptable Protocol Version"
	case ConnectIdentifierRejected:
		return "Identifier Rejected"
	case ConnectServerUnavailable:
		return "Server Unavailable"
	case ConnectMalformedUsernameOrPassword:
		return "Malformed Username or Password"
	case ConnectNotAuthorized:
		return "Not Authorized"
	}
	return "Unknown"
}

func (c ConnectReturnCode) valid() bool {
	switch c {
	case ConnectAccepted:
	case ConnectUnacceptableProtocolVersion:
	case ConnectIdentifierRejected:
	case ConnectServerUnavailable:
	case ConnectMalformedUsernameOrPassword:
	case ConnectNotAuthorized:
	default:
		return false
	}
	return true
}
