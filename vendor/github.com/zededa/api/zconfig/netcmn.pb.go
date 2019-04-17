// Code generated by protoc-gen-go. DO NOT EDIT.
// source: netcmn.proto

package zconfig

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ProxyProto int32

const (
	ProxyProto_PROXY_HTTP  ProxyProto = 0
	ProxyProto_PROXY_HTTPS ProxyProto = 1
	ProxyProto_PROXY_SOCKS ProxyProto = 2
	ProxyProto_PROXY_FTP   ProxyProto = 3
	ProxyProto_PROXY_OTHER ProxyProto = 255
)

var ProxyProto_name = map[int32]string{
	0:   "PROXY_HTTP",
	1:   "PROXY_HTTPS",
	2:   "PROXY_SOCKS",
	3:   "PROXY_FTP",
	255: "PROXY_OTHER",
}
var ProxyProto_value = map[string]int32{
	"PROXY_HTTP":  0,
	"PROXY_HTTPS": 1,
	"PROXY_SOCKS": 2,
	"PROXY_FTP":   3,
	"PROXY_OTHER": 255,
}

func (x ProxyProto) String() string {
	return proto.EnumName(ProxyProto_name, int32(x))
}
func (ProxyProto) EnumDescriptor() ([]byte, []int) { return fileDescriptor9, []int{0} }

type DHCPType int32

const (
	DHCPType_DHCPNoop DHCPType = 0
	// Statically configure the DHCP for port
	DHCPType_Static DHCPType = 1
	// Don't run any DHCP, we are in passthrough mode for app
	DHCPType_DHCPNone DHCPType = 2
	// Deprecated Server no longer
	DHCPType_deprecated_server DHCPType = 3
	// Run the DHCP client on this port
	DHCPType_Client DHCPType = 4
)

var DHCPType_name = map[int32]string{
	0: "DHCPNoop",
	1: "Static",
	2: "DHCPNone",
	3: "deprecated_server",
	4: "Client",
}
var DHCPType_value = map[string]int32{
	"DHCPNoop":          0,
	"Static":            1,
	"DHCPNone":          2,
	"deprecated_server": 3,
	"Client":            4,
}

func (x DHCPType) String() string {
	return proto.EnumName(DHCPType_name, int32(x))
}
func (DHCPType) EnumDescriptor() ([]byte, []int) { return fileDescriptor9, []int{1} }

type NetworkType int32

const (
	NetworkType_NETWORKTYPENOOP NetworkType = 0
	NetworkType_V4              NetworkType = 4
	NetworkType_V6              NetworkType = 6
	NetworkType_CryptoV4        NetworkType = 24
	NetworkType_CryptoV6        NetworkType = 26
	NetworkType_CryptoEID       NetworkType = 14
)

var NetworkType_name = map[int32]string{
	0:  "NETWORKTYPENOOP",
	4:  "V4",
	6:  "V6",
	24: "CryptoV4",
	26: "CryptoV6",
	14: "CryptoEID",
}
var NetworkType_value = map[string]int32{
	"NETWORKTYPENOOP": 0,
	"V4":              4,
	"V6":              6,
	"CryptoV4":        24,
	"CryptoV6":        26,
	"CryptoEID":       14,
}

func (x NetworkType) String() string {
	return proto.EnumName(NetworkType_name, int32(x))
}
func (NetworkType) EnumDescriptor() ([]byte, []int) { return fileDescriptor9, []int{2} }

type IpRange struct {
	Start string `protobuf:"bytes,1,opt,name=start" json:"start,omitempty"`
	End   string `protobuf:"bytes,2,opt,name=end" json:"end,omitempty"`
}

func (m *IpRange) Reset()                    { *m = IpRange{} }
func (m *IpRange) String() string            { return proto.CompactTextString(m) }
func (*IpRange) ProtoMessage()               {}
func (*IpRange) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{0} }

func (m *IpRange) GetStart() string {
	if m != nil {
		return m.Start
	}
	return ""
}

func (m *IpRange) GetEnd() string {
	if m != nil {
		return m.End
	}
	return ""
}

type ProxyServer struct {
	Proto  ProxyProto `protobuf:"varint,1,opt,name=proto,enum=ProxyProto" json:"proto,omitempty"`
	Server string     `protobuf:"bytes,2,opt,name=server" json:"server,omitempty"`
	Port   uint32     `protobuf:"varint,3,opt,name=port" json:"port,omitempty"`
}

func (m *ProxyServer) Reset()                    { *m = ProxyServer{} }
func (m *ProxyServer) String() string            { return proto.CompactTextString(m) }
func (*ProxyServer) ProtoMessage()               {}
func (*ProxyServer) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{1} }

func (m *ProxyServer) GetProto() ProxyProto {
	if m != nil {
		return m.Proto
	}
	return ProxyProto_PROXY_HTTP
}

func (m *ProxyServer) GetServer() string {
	if m != nil {
		return m.Server
	}
	return ""
}

func (m *ProxyServer) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

type ProxyConfig struct {
	// enable network level proxy in the form of WPAD
	NetworkProxyEnable bool `protobuf:"varint,1,opt,name=networkProxyEnable" json:"networkProxyEnable,omitempty"`
	// dedicated per protocol information
	Proxies []*ProxyServer `protobuf:"bytes,2,rep,name=proxies" json:"proxies,omitempty"`
	// exceptions seperated by commas
	Exceptions string `protobuf:"bytes,3,opt,name=exceptions" json:"exceptions,omitempty"`
	// or pacfile can be in place of others
	// base64 encoded
	Pacfile string `protobuf:"bytes,4,opt,name=pacfile" json:"pacfile,omitempty"`
	// Direct URL for wpad.dat download
	NetworkProxyURL string `protobuf:"bytes,5,opt,name=networkProxyURL" json:"networkProxyURL,omitempty"`
}

func (m *ProxyConfig) Reset()                    { *m = ProxyConfig{} }
func (m *ProxyConfig) String() string            { return proto.CompactTextString(m) }
func (*ProxyConfig) ProtoMessage()               {}
func (*ProxyConfig) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{2} }

func (m *ProxyConfig) GetNetworkProxyEnable() bool {
	if m != nil {
		return m.NetworkProxyEnable
	}
	return false
}

func (m *ProxyConfig) GetProxies() []*ProxyServer {
	if m != nil {
		return m.Proxies
	}
	return nil
}

func (m *ProxyConfig) GetExceptions() string {
	if m != nil {
		return m.Exceptions
	}
	return ""
}

func (m *ProxyConfig) GetPacfile() string {
	if m != nil {
		return m.Pacfile
	}
	return ""
}

func (m *ProxyConfig) GetNetworkProxyURL() string {
	if m != nil {
		return m.NetworkProxyURL
	}
	return ""
}

// These are list of static mapping that can be added to network
type ZnetStaticDNSEntry struct {
	HostName string   `protobuf:"bytes,1,opt,name=HostName" json:"HostName,omitempty"`
	Address  []string `protobuf:"bytes,2,rep,name=Address" json:"Address,omitempty"`
}

func (m *ZnetStaticDNSEntry) Reset()                    { *m = ZnetStaticDNSEntry{} }
func (m *ZnetStaticDNSEntry) String() string            { return proto.CompactTextString(m) }
func (*ZnetStaticDNSEntry) ProtoMessage()               {}
func (*ZnetStaticDNSEntry) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{3} }

func (m *ZnetStaticDNSEntry) GetHostName() string {
	if m != nil {
		return m.HostName
	}
	return ""
}

func (m *ZnetStaticDNSEntry) GetAddress() []string {
	if m != nil {
		return m.Address
	}
	return nil
}

// Common for IPv4 and IPv6
type Ipspec struct {
	Dhcp DHCPType `protobuf:"varint,2,opt,name=dhcp,enum=DHCPType" json:"dhcp,omitempty"`
	// subnet is CIDR format...x.y.z.l/nn
	Subnet  string   `protobuf:"bytes,3,opt,name=subnet" json:"subnet,omitempty"`
	Gateway string   `protobuf:"bytes,5,opt,name=gateway" json:"gateway,omitempty"`
	Domain  string   `protobuf:"bytes,6,opt,name=domain" json:"domain,omitempty"`
	Ntp     string   `protobuf:"bytes,7,opt,name=ntp" json:"ntp,omitempty"`
	Dns     []string `protobuf:"bytes,8,rep,name=dns" json:"dns,omitempty"`
	// for IPAM management when dhcp is turned on.
	// If none provided, system will default pool.
	DhcpRange *IpRange `protobuf:"bytes,9,opt,name=dhcpRange" json:"dhcpRange,omitempty"`
}

func (m *Ipspec) Reset()                    { *m = Ipspec{} }
func (m *Ipspec) String() string            { return proto.CompactTextString(m) }
func (*Ipspec) ProtoMessage()               {}
func (*Ipspec) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{4} }

func (m *Ipspec) GetDhcp() DHCPType {
	if m != nil {
		return m.Dhcp
	}
	return DHCPType_DHCPNoop
}

func (m *Ipspec) GetSubnet() string {
	if m != nil {
		return m.Subnet
	}
	return ""
}

func (m *Ipspec) GetGateway() string {
	if m != nil {
		return m.Gateway
	}
	return ""
}

func (m *Ipspec) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *Ipspec) GetNtp() string {
	if m != nil {
		return m.Ntp
	}
	return ""
}

func (m *Ipspec) GetDns() []string {
	if m != nil {
		return m.Dns
	}
	return nil
}

func (m *Ipspec) GetDhcpRange() *IpRange {
	if m != nil {
		return m.DhcpRange
	}
	return nil
}

func init() {
	proto.RegisterType((*IpRange)(nil), "ipRange")
	proto.RegisterType((*ProxyServer)(nil), "ProxyServer")
	proto.RegisterType((*ProxyConfig)(nil), "ProxyConfig")
	proto.RegisterType((*ZnetStaticDNSEntry)(nil), "ZnetStaticDNSEntry")
	proto.RegisterType((*Ipspec)(nil), "ipspec")
	proto.RegisterEnum("ProxyProto", ProxyProto_name, ProxyProto_value)
	proto.RegisterEnum("DHCPType", DHCPType_name, DHCPType_value)
	proto.RegisterEnum("NetworkType", NetworkType_name, NetworkType_value)
}

func init() { proto.RegisterFile("netcmn.proto", fileDescriptor9) }

var fileDescriptor9 = []byte{
	// 606 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0xdd, 0x6e, 0xda, 0x4a,
	0x10, 0x8e, 0x81, 0x00, 0x1e, 0x12, 0xb2, 0x67, 0xcf, 0x39, 0x95, 0x15, 0x29, 0x2d, 0xe5, 0x22,
	0x42, 0xb9, 0x70, 0x54, 0x1a, 0xe5, 0xba, 0x2d, 0xa1, 0xa2, 0x4d, 0x05, 0xd6, 0xe2, 0xa4, 0x4d,
	0xd4, 0x2a, 0x32, 0xf6, 0x84, 0x58, 0x85, 0xdd, 0x95, 0xbd, 0x34, 0x21, 0xef, 0xd6, 0x07, 0xe8,
	0x53, 0xb5, 0xda, 0x5d, 0x53, 0x50, 0xd5, 0x2b, 0xcf, 0xf7, 0xcd, 0xdf, 0xe7, 0x99, 0x59, 0xd8,
	0xe1, 0xa8, 0xe2, 0x39, 0xf7, 0x65, 0x26, 0x94, 0x68, 0xbf, 0x80, 0x5a, 0x2a, 0x59, 0xc4, 0xa7,
	0x48, 0xff, 0x83, 0xed, 0x5c, 0x45, 0x99, 0xf2, 0x9c, 0x96, 0xd3, 0x71, 0x99, 0x05, 0x94, 0x40,
	0x19, 0x79, 0xe2, 0x95, 0x0c, 0xa7, 0xcd, 0xf6, 0x67, 0x68, 0x04, 0x99, 0x78, 0x58, 0x8e, 0x31,
	0xfb, 0x86, 0x19, 0x7d, 0x0e, 0xdb, 0xa6, 0x94, 0x49, 0x6b, 0x76, 0x1b, 0xba, 0xf0, 0xc3, 0x32,
	0xd0, 0x14, 0xb3, 0x1e, 0xfa, 0x04, 0xaa, 0xb9, 0x09, 0x2e, 0xca, 0x14, 0x88, 0x52, 0xa8, 0x48,
	0x91, 0x29, 0xaf, 0xdc, 0x72, 0x3a, 0xbb, 0xcc, 0xd8, 0xed, 0x1f, 0x4e, 0x51, 0xbe, 0x27, 0xf8,
	0x6d, 0x3a, 0xa5, 0x3e, 0x50, 0x8e, 0xea, 0x5e, 0x64, 0x5f, 0x0d, 0xdb, 0xe7, 0xd1, 0x64, 0x86,
	0xa6, 0x57, 0x9d, 0xfd, 0xc5, 0x43, 0x0f, 0xa1, 0xa6, 0x05, 0xa4, 0x98, 0x7b, 0xa5, 0x56, 0xb9,
	0xd3, 0xe8, 0xee, 0xf8, 0x1b, 0x6a, 0xd9, 0xca, 0x49, 0x9f, 0x02, 0xe0, 0x43, 0x8c, 0x52, 0xa5,
	0x82, 0xe7, 0x46, 0x81, 0xcb, 0x36, 0x18, 0xea, 0x41, 0x4d, 0x46, 0xf1, 0x6d, 0x3a, 0x43, 0xaf,
	0x62, 0x9c, 0x2b, 0x48, 0x3b, 0xb0, 0xb7, 0xd9, 0xf7, 0x82, 0x7d, 0xf0, 0xb6, 0x4d, 0xc4, 0x9f,
	0x74, 0xfb, 0x3d, 0xd0, 0x6b, 0x8e, 0x6a, 0xac, 0x22, 0x95, 0xc6, 0x67, 0xc3, 0x71, 0x9f, 0xab,
	0x6c, 0x49, 0xf7, 0xa1, 0x3e, 0x10, 0xb9, 0x1a, 0x46, 0x73, 0x2c, 0x46, 0xfd, 0x1b, 0xeb, 0xae,
	0xaf, 0x93, 0x24, 0xc3, 0xdc, 0xaa, 0x77, 0xd9, 0x0a, 0xb6, 0xbf, 0x3b, 0x50, 0x4d, 0x65, 0x2e,
	0x31, 0xa6, 0x07, 0x50, 0x49, 0xee, 0x62, 0x69, 0x86, 0xd9, 0xec, 0xba, 0xfe, 0xd9, 0xa0, 0x17,
	0x84, 0x4b, 0x89, 0xcc, 0xd0, 0x66, 0xda, 0x8b, 0x09, 0x47, 0x55, 0xfc, 0x55, 0x81, 0x74, 0xed,
	0x69, 0xa4, 0xf0, 0x3e, 0x5a, 0x16, 0x7a, 0x57, 0x50, 0x67, 0x24, 0x62, 0x1e, 0xa5, 0xdc, 0xab,
	0xda, 0x0c, 0x8b, 0xf4, 0xee, 0xb9, 0x92, 0x5e, 0xcd, 0xee, 0x9e, 0x2b, 0xa9, 0x99, 0x84, 0xe7,
	0x5e, 0xdd, 0x68, 0xd3, 0x26, 0x3d, 0x04, 0x57, 0x77, 0x35, 0x27, 0xe4, 0xb9, 0x2d, 0xa7, 0xd3,
	0xe8, 0xd6, 0xfd, 0xe2, 0xa4, 0xd8, 0xda, 0x75, 0x74, 0x03, 0xb0, 0x3e, 0x0c, 0xda, 0x04, 0x08,
	0xd8, 0xe8, 0xd3, 0xd5, 0xcd, 0x20, 0x0c, 0x03, 0xb2, 0x45, 0xf7, 0xa0, 0xb1, 0xc6, 0x63, 0xe2,
	0xac, 0x89, 0xf1, 0xa8, 0x77, 0x3e, 0x26, 0x25, 0xba, 0x0b, 0xae, 0x25, 0xde, 0x86, 0x01, 0x29,
	0x53, 0xb2, 0xf2, 0x8f, 0xc2, 0x41, 0x9f, 0x91, 0x9f, 0xce, 0xd1, 0x05, 0xd4, 0x57, 0x83, 0xa0,
	0x3b, 0xd6, 0x1e, 0x0a, 0x21, 0xc9, 0x16, 0x05, 0xa8, 0xda, 0x15, 0x10, 0x67, 0xed, 0xe1, 0x48,
	0x4a, 0xf4, 0x7f, 0xf8, 0x27, 0x41, 0x99, 0x61, 0x1c, 0x29, 0x4c, 0x6e, 0xec, 0x55, 0x92, 0xb2,
	0x4e, 0xe8, 0xcd, 0x52, 0xe4, 0x8a, 0x54, 0x8e, 0xbe, 0x40, 0x63, 0x68, 0xd7, 0x6a, 0x2a, 0xff,
	0x0b, 0x7b, 0xc3, 0x7e, 0xf8, 0x71, 0xc4, 0xce, 0xc3, 0xab, 0xa0, 0x3f, 0x1c, 0x8d, 0xb4, 0xfa,
	0x2a, 0x94, 0x2e, 0x4f, 0x48, 0xc5, 0x7c, 0x4f, 0x49, 0x55, 0x37, 0xe9, 0x65, 0x4b, 0xa9, 0xc4,
	0xe5, 0x09, 0xf1, 0x36, 0xd0, 0x29, 0xd9, 0xd7, 0xff, 0x61, 0x51, 0xff, 0xdd, 0x19, 0x69, 0xbe,
	0x79, 0x05, 0xcf, 0x62, 0x31, 0xf7, 0x1f, 0x31, 0xc1, 0x24, 0xf2, 0xe3, 0x99, 0x58, 0x24, 0xfe,
	0x42, 0x0b, 0x49, 0x63, 0xb4, 0x4f, 0xf4, 0xfa, 0x60, 0x9a, 0xaa, 0xbb, 0xc5, 0xc4, 0x8f, 0xc5,
	0xfc, 0xd8, 0xc6, 0x1d, 0x47, 0x32, 0x3d, 0x7e, 0x8c, 0xcd, 0x03, 0x99, 0x54, 0x4d, 0xd4, 0xcb,
	0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xb8, 0x5f, 0x81, 0x5e, 0xd8, 0x03, 0x00, 0x00,
}
