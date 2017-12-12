package itr

import (
	"fmt"
	"net"
	"syscall"
	"time"
	//"sync"
	"encoding/json"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pfring"
	"github.com/zededa/go-provision/dataplane/fib"
	"github.com/zededa/go-provision/types"
	"os"
)

const SNAPLENGTH = 65536

func StartItrThread(threadName string,
	killChannel chan bool,
	puntChannel chan []byte) {

	fmt.Println("Starting thread:", threadName)
	// Kill channel will no longer be needed
	// if we return from this function
	defer close(killChannel)

	ring := setupPacketCapture(threadName, SNAPLENGTH)
	if ring == nil {
		fmt.Fprintf(os.Stderr, "Packet capture setup for interface %s failed\n",
			threadName)
	}
	defer ring.Close()

	// create raw socket pair for sending LISP packets out
	fd4, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed creating IPv4 raw socket for %s: %s\n",
			threadName, err)
		return
	}
	defer syscall.Close(fd4)
	fd6, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed creating IPv6 raw socket for %s: %s\n",
			threadName, err)
		return
	}
	defer syscall.Close(fd6)

	startWorking(threadName, ring, killChannel, puntChannel, fd4, fd6)

	// If startWorking returns, it means the control thread wants
	// this thread to die.
	return
}

func setupPacketCapture(ifname string, snapLen uint32) *pfring.Ring {
	// create a new pf_ring to capture packets from our interface
	ring, err := pfring.NewRing(ifname, SNAPLENGTH, pfring.FlagPromisc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "PF_RING creation for interface %s failed: %s\n",
			ifname, err)
		return nil
	}

	// Capture ipv6 packets only
	err = ring.SetBPFFilter("ip6")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Setting ipv6 BPF filter on interface %s failed: %s\n",
			ifname, err)
		ring.Close()
		return nil
	}

	// Make PF_RING capture only transmitted packet
	ring.SetDirection(pfring.TransmitOnly)

	// set the ring in readonly mode
	ring.SetSocketMode(pfring.ReadOnly)

	// Enable ring. Packet inflow starts after this.
	err = ring.Enable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed enabling PF_RING for interface %s: %s\n",
			ifname, err)
		ring.Close()
		return nil
	}
	return ring
}

// Start capturing and processing packets.
func startWorking(ifname string, ring *pfring.Ring,
	killChannel chan bool, puntChannel chan []byte,
	//conn4 net.PacketConn, conn6 net.PacketConn) {
	fd4 int, fd6 int) {
	var pktBuf [SNAPLENGTH]byte

	iid := fib.LookupIfaceIID(ifname)
	if iid == 0 {
		fmt.Fprintf(os.Stderr, "Interface %s's IID cannot be found\n", ifname)
		return
	}

	// We need the EIDs attached to this interfaces for further processing
	// Keep looking for them every 100ms
	var eids []net.IP
eidLoop:
	for {
		time.Sleep(2 * time.Second)
		select {
		case <-killChannel:
			fmt.Printf("ITR thread %s received terminate from control module.", ifname)
			return
		default:
			// EID map database might not have come yet. Wait for before we start
			// processing packets.
			eids = fib.LookupIfaceEids(iid)
			if eids != nil {
				break eidLoop
			}
			fmt.Println("Re-trying EID lookup for interface", ifname)
			continue
		}
	}

	/*
	 * While waiting for packets we should also look for the terminate
	 * message from control module. If control module sends a terminate,
	 * ITR thread should free all its allocated resources, stop processing
	 * packets and exit.
	 */
	for {
		select {
		case <-killChannel:
			fmt.Printf("ITR thread %s received terminate from control module.", ifname)
			return
		default:
			/*
				ci, err := ring.ReadPacketDataToNoWait(pktBuf[:])
				if err == pfring.NextNoPacketNonblocking {
					//fmt.Println("No packet, socket is non blocking")
					continue
				}
			*/
			ci, err := ring.ReadPacketDataTo(pktBuf[fib.MAXHEADERLEN:])
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Something wrong with packet capture from interface %s: %s\n",
					ifname, err)
				return
			}

			pktLen := ci.CaptureLength
			if pktLen <= 0 {
				// XXX May be add a per thread stat here
				continue
			}
			packet := gopacket.NewPacket(pktBuf[fib.MAXHEADERLEN:ci.CaptureLength+fib.MAXHEADERLEN],
				layers.LinkTypeEthernet,
				gopacket.DecodeOptions{Lazy: true, NoCopy: true})
			ip6Layer := packet.Layer(layers.LayerTypeIPv6)
			if ip6Layer == nil {
				// XXX May be add a per thread stat here
				// Ignore this packet.
				continue
			}

			ipHeader := ip6Layer.(*layers.IPv6)

			// Check if the source address of packet matches with any of the eids
			// assigned to input interface.
			srcAddr := ipHeader.SrcIP
			matchFound := false
			for _, eid := range eids {
				if srcAddr.Equal(eid) == true {
					matchFound = true
					break
				}
			}

			if !matchFound {
				// XXX May be add a per thread stat here
				fmt.Fprintf(os.Stderr,
				"Input packet with source address %s does not have matching EID of interface\n",
				srcAddr)
				continue
			}

			dstAddr := ipHeader.DstIP

			/**
			 * Compute hash of packet.
			 * LSB 4 bytes of src addr (xor) LSB 4 bytes of dst addr (xor)
			 * (src port << 16 | dst port)
			 */
			var srcAddrBytes uint32 = (uint32(srcAddr[12])<<24 |
				uint32(srcAddr[13])<<16 |
				uint32(srcAddr[14])<<8 | uint32(srcAddr[15]))
			var dstAddrBytes uint32 = (uint32(dstAddr[12])<<24 |
				uint32(dstAddr[13])<<16 |
				uint32(dstAddr[14])<<8 | uint32(dstAddr[15]))
			transportLayer := packet.TransportLayer()

			// This is a byte array of the header
			transportContents := transportLayer.LayerContents()
			var ports uint32 = (uint32(transportContents[0])<<24 |
				uint32(transportContents[1])<<16 |
				uint32(transportContents[2])<<8 |
				uint32(transportContents[3]))

			var hash32 uint32 = srcAddrBytes ^ dstAddrBytes ^ ports

			/*
				fmt.Println("srcAddrBytes:", srcAddrBytes)
				fmt.Println("dstAddrBytes:", dstAddrBytes)
				fmt.Println("ports:", ports)
				fmt.Println("hash32:", hash32)
			*/

			LookupAndSend(packet, pktBuf[:],
				uint32(pktLen), iid, hash32,
				ifname, srcAddr, dstAddr,
				//puntChannel, conn4, conn6)
				puntChannel, fd4, fd6)
		}
	}
}

// This function expectes the parameter pktBuf to be a statically
// allocated buffer longer than the original packet length.
// We currently use a buffer of length 64K bytes.
//
// Perform lookup into mapcache database and forward if the lookup succeeds.
// If not, buffer the packet and send a punt request to lispers.net for resolution.
// Look for comments inside the function to understand more about what it does.
func LookupAndSend(packet gopacket.Packet,
	pktBuf []byte,
	capLen uint32,
	iid uint32,
	hash32 uint32,
	ifname string,
	srcAddr net.IP,
	dstAddr net.IP,
	puntChannel chan []byte,
	//conn4 net.PacketConn, conn6 net.PacketConn) {
	fd4 int, fd6 int) {
	mapEntry, punt := fib.LookupAndAdd(iid, dstAddr)
	if mapEntry.Resolved != true {
		// Buffer the packet for processing later

		// Add packet to channel in a non blocking fashion.
		// Buffered packet channel is only 10 entries long.
		select {
		case mapEntry.PktBuffer <- &types.BufferedPacket{
			Packet: packet,
			Hash32: hash32,
		}:
		default:
			fmt.Println("Packet buffer channel full for EID", dstAddr)
		}

		/**
		 * There is no guarantee that the control thread has not
		 * resolved our unresolved map entry by the time we add packet
		 * to buffered packet channel. We perform the lookup for our
		 * iid, eid once again with read lock and check the resolution
		 * status.
		 *
		 * If the map cache entry is resolved by now, we dequeue one
		 * packet from the buffered packet channel and send it out.
		 * This avoids the case where control thread has already sent
		 * out all buffered packets and our packet sits in the buffered
		 * channel without being noticed.
		 */
		mapEntry, punt = fib.LookupAndAdd(iid, dstAddr)
		if mapEntry.Resolved {
			select {
			case pkt := <-mapEntry.PktBuffer:
				// XXX Send this packet out
				// Packet read is still in pktBuf buffer. It is safe to pass it's
				// pointer.
				fib.CraftAndSendLispPacket(pkt.Packet, pktBuf, capLen, pkt.Hash32,
					mapEntry, iid, fd4, fd6)
				//iid, conn4, conn6)
			default:
				// We do not want to get blocked and keep waiting
				// when there are no packets in the buffer channel.
			}
		}
	} else {
		// Craft the LISP header, outer layers here and send packet out
		fib.CraftAndSendLispPacket(packet, pktBuf, capLen, hash32, mapEntry,
			iid, fd4, fd6)
		//iid, conn4, conn6)
	}
	if punt == true {
		// We will have to put a punt request on the control
		// module's channel
		puntEntry := types.PuntEntry{
			Deid:  dstAddr,
			Seid:  srcAddr,
			Iface: ifname,
		}
		puntMsg, err := json.Marshal(puntEntry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Marshaling punt entry failed %s: %s\n",
				puntEntry, err)
		} else {
			puntChannel <- puntMsg
			fmt.Println("Sending punt entry at", time.Now())
		}
	}
	return
}
