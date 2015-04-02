package ssh

import (
	"encoding/binary"
	"fmt"
	"regexp"

	"github.com/zmap/zgrab/ztools/zlog"
)

type ProtocolVersion string

// ProtocolAgreement represents the client and server protocol banners
//
// RFC specifies the format for the banner as specified in RFC 4523 Section 4.2
//       SSH-protoversion-softwareversion SP comments CR LF
//
// The server MAY send other lines of data before sending the version.
// The lines are terminated by CR LF, and SHOULD be encoded as UTF-8
//
// See http://tools.ietf.org/html/rfc4253 for details
type ProtocolAgreement struct {
	RawBanner       string `json:"raw_banner,omitempty"`
	ProtocolVersion string `json:"protocol_version,omitempty"`
	SoftwareVersion string `json:"software_version,omitempty"`
	Comments        string `json:"comments,omitempty"`
}

// MakeZGrabProtocolAgreement returns the default client
// ProtocolAgreement message for ZGrab.
//
// This sets protocol version to "2.0", software version to "ZGrab",
// and comments to "ZGrab SSH Survey"
func MakeZGrabProtocolAgreement() *ProtocolAgreement {
	h := ProtocolAgreement{
		ProtocolVersion: "2.0",
		SoftwareVersion: "ZGrab",
		Comments:        "ZGrab SSH Survey",
	}
	return &h
}

// ParseRawBanner populates a ProtocolAgreement struct based on the
// contents of the RawBanner field.
func (h *ProtocolAgreement) ParseRawBanner() {
	matches := serverBannerRegex.FindStringSubmatch(h.RawBanner)
	if len(matches) >= 2 {
		h.ProtocolVersion = matches[1]
	}
	if len(matches) >= 3 {
		h.SoftwareVersion = matches[2]
	}
	if len(matches) >= 4 {
		h.Comments = matches[3]
	}
}

// Marshal returns a byte array suitable for a call to write
func (h *ProtocolAgreement) Marshal() []byte {
	if h.RawBanner != "" {
		return []byte(h.RawBanner)
	}
	if h.Comments != "" {
		h.RawBanner = fmt.Sprintf("SSH-%s-%s %s\r\n", h.ProtocolVersion, h.SoftwareVersion, h.Comments)
	} else {
		h.RawBanner = fmt.Sprintf("SSH-%s-%s\r\n", h.ProtocolVersion, h.SoftwareVersion)
	}
	return []byte(h.RawBanner)
}

var serverBannerRegex = regexp.MustCompile(`(?:SSH-([!-,.-~]*)-([!-,.-~]*) ([^\r\n]*)\r\n)|(?:SSH-([!-,.-~]*)-([!-,.-~]*)\r\n)`)

// The Cookie type represents the random cookie sent during the server
// key exchange.
type Cookie [16]byte

/*
// MarshalJSON encodes a Cookie to JSON as a base64-encoded byte array.
func (c *Cookie) MarshalJSON() ([]byte, error) {
	return json.Marshal(c[:])
}

// UnmarshalJSON unmarshal a byte-array encoded in base64 to a cookie.
// The byte array must be either null, empty array, or exactly 16 bytes
// long.
func (c *Cookie) UnmarshalJSON(b []byte) error {
	raw := c[:]
	if err := json.Unmarshal(b, raw); err != nil {
		return err
	}
	length := len(raw)
	if length != 0 && length != 16 {
		return fmt.Errorf("Cookies must be 16 bytes long, given %d", length)
	}
	return nil
}
*/

type KeyExchangeInit struct {
	raw                       []byte
	Cookie                    Cookie   `json:"cookie"`
	KexAlgorithms             NameList `json:"key_exchange_algorithms"`
	HostKeyAlgorithms         NameList `json:"host_key_algorithms"`
	EncryptionClientToServer  NameList `json:"encryption_client_to_server"`
	EncryptionServerToClient  NameList `json:"encryption_server_to_client"`
	MACClientToServer         NameList `json:"mac_client_to_server"`
	MACServerToClient         NameList `json:"mac_server_to_client"`
	CompressionClientToServer NameList `json:"compression_client_to_server"`
	CompressionServerToClient NameList `json:"compression_server_to_client"`
	LanguageClientToServer    NameList `json:"language_client_to_server"`
	LanguageServerToClient    NameList `json:"language_server_to_client"`
	FirstKexPacketFollows     bool     `json:"first_kex_packet_follows"`
	Zero                      uint32   `json:"zero"`
}

func (kxi *KeyExchangeInit) Marshal() ([]byte, error) {
	if kxi.raw != nil {
		return kxi.raw, nil
	}
	payloadLength := 0
	payloadLength += 16 // Cookie
	payloadLength += kxi.KexAlgorithms.MarshaledLength()
	payloadLength += kxi.HostKeyAlgorithms.MarshaledLength()
	payloadLength += kxi.EncryptionClientToServer.MarshaledLength()
	payloadLength += kxi.EncryptionServerToClient.MarshaledLength()
	payloadLength += kxi.MACClientToServer.MarshaledLength()
	payloadLength += kxi.MACServerToClient.MarshaledLength()
	payloadLength += kxi.CompressionClientToServer.MarshaledLength()
	payloadLength += kxi.CompressionServerToClient.MarshaledLength()
	payloadLength += kxi.LanguageClientToServer.MarshaledLength()
	payloadLength += kxi.LanguageServerToClient.MarshaledLength()
	payloadLength += 1 + 4 // Bool + Reserved
	out := make([]byte, payloadLength)
	b := out
	copy(b[0:16], kxi.Cookie[:])
	b = b[16:]
	var err error
	if b, err = kxi.KexAlgorithms.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.HostKeyAlgorithms.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.EncryptionClientToServer.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.EncryptionServerToClient.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.MACClientToServer.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.MACServerToClient.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.CompressionClientToServer.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.CompressionServerToClient.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.LanguageClientToServer.MarshalInto(b); err != nil {
		return nil, err
	}
	if b, err = kxi.LanguageServerToClient.MarshalInto(b); err != nil {
		return nil, err
	}
	return out, nil
}

// Unmarshal a byte array into a KeyExchangeInit struct. The byte array
// should be the entire KeyExchangeInit message payload starting after
// the SSH_MSG_KEXINIT byte.
func (kxi *KeyExchangeInit) Unmarshal(raw []byte) bool {
	kxi.raw = raw
	b := raw
	if len(b) < 16 {
		return false
	}
	copy(kxi.Cookie[:], b[0:16])
	b = b[16:]
	var ok bool
	if b, ok = kxi.KexAlgorithms.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.HostKeyAlgorithms.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.EncryptionClientToServer.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.EncryptionServerToClient.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.MACClientToServer.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.MACServerToClient.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.CompressionClientToServer.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.CompressionServerToClient.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.LanguageClientToServer.Unmarshal(b); !ok {
		return false
	}
	if b, ok = kxi.LanguageServerToClient.Unmarshal(b); !ok {
		return false
	}
	if len(b) < 1 {
		return false
	}
	zlog.Debug(b)

	kxi.FirstKexPacketFollows = (b[0] != 0)
	b = b[1:]
	if len(b) < 4 {
		return false
	}

	// Read the zero out from the end
	kxi.Zero = binary.BigEndian.Uint32(b)
	b = b[4:]

	// We should be done
	if len(b) != 0 {
		zlog.Debug("WAAAT")
		zlog.Debug(len(b))
		return false
	}
	return true
}