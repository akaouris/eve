// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

// Create ip rules and ip routing tables for each ifindex and also a free
// one for the collection of free uplinks.

package zedrouter

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"syscall"
)

var FreeTable = 500 // Need a FreeUplink policy for NAT+underlay

type addrChangeFnType func(ifname string)
type suppressRoutesFnType func(ifname string) bool

// XXX should really be in a context returned by Init
var addrChangeFuncUplink addrChangeFnType
var addrChangeFuncNonUplink addrChangeFnType
var suppressRoutesFunc suppressRoutesFnType

// Returns the channels for route, addr, link updates
func PbrInit(uplinks []string, freeUplinks []string,
	addrChange addrChangeFnType,
	addrChangeNon addrChangeFnType,
	suppressRoutes suppressRoutesFnType) (chan netlink.RouteUpdate,
	chan netlink.AddrUpdate, chan netlink.LinkUpdate) {

	if debug {
		log.Printf("PbrInit(%v, %v)\n", uplinks, freeUplinks)
	}
	setUplinks(uplinks)
	setFreeUplinks(freeUplinks)
	addrChangeFuncUplink = addrChange
	addrChangeFuncNonUplink = addrChangeNon
	suppressRoutesFunc = suppressRoutes

	IfindexToNameInit()
	IfindexToAddrsInit()

	flushRoutesTable(FreeTable, 0)

	// flush any old rules using RuleList
	flushRules(0)

	// Create rule for FreeTable; src NAT range
	// XXX for IPv6 underlay we also need rules.
	// Can we use iif match for all the bo* interfaces?
	// If so, use bu* matches for this rule
	freeRule := netlink.NewRule()
	// XXX need this rule for all NAT subnets
	_, prefix, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		log.Fatal(err)
	}
	freeRule.Src = prefix
	freeRule.Table = FreeTable
	freeRule.Family = syscall.AF_INET
	// Avoid duplicate rules
	_ = netlink.RuleDel(freeRule)
	err = netlink.RuleAdd(freeRule)
	if err != nil {
		log.Fatal(err)
	}

	// Need links to get name to ifindex? Or lookup each time?
	linkchan := make(chan netlink.LinkUpdate)
	linkopt := netlink.LinkSubscribeOptions{ListExisting: true}
	if err := netlink.LinkSubscribeWithOptions(linkchan, nil,
		linkopt); err != nil {
		log.Fatal(err)
	}

	addrchan := make(chan netlink.AddrUpdate)
	addropt := netlink.AddrSubscribeOptions{ListExisting: true}
	if err := netlink.AddrSubscribeWithOptions(addrchan, nil,
		addropt); err != nil {
		log.Fatal(err)
	}
	routechan := make(chan netlink.RouteUpdate)
	rtopt := netlink.RouteSubscribeOptions{ListExisting: true}
	if err := netlink.RouteSubscribeWithOptions(routechan, nil,
		rtopt); err != nil {
		log.Fatal(err)
	}
	return routechan, addrchan, linkchan
}

// Handle a route change
func PbrRouteChange(change netlink.RouteUpdate) {
	rt := change.Route
	if rt.Table != syscall.RT_TABLE_MAIN {
		// Ignore since we will not add to other table
		return
	}
	doFreeTable := false
	ifname, err := IfindexToName(rt.LinkIndex)
	if err != nil {
		// We'll check on ifname when we see a linkchange
		log.Printf("PbrRouteChange IfindexToName failed for %d: %s\n",
			rt.LinkIndex, err)
	} else {
		if isFreeUplink(ifname) {
			if debug {
				log.Printf("Applying to FreeTable: %v\n", rt)
			}
			doFreeTable = true
		}
		if suppressRoutesFunc != nil &&
			suppressRoutesFunc(ifname) {
			// Delete any route which was added on an assignable
			// adapter.
			// XXX alternative is to add it with a very high metric
			// but that requires a delete and re-add since the
			// metric can't be changed.
			// XXX needs work to handle adding a link as an uplink
			// on a running system, but perhaps that will only be
			// done using a reboot
			if err := netlink.RouteDel(&rt); err != nil {
				// XXX Fatal?
				log.Printf("PrbRouteChange - RouteDel %v failed %s\n",
					rt, err)
			} else {
				log.Printf("PrbRouteChange suppress RouteDel %v\n",
					rt)
			}
			return
		}
	}
	srt := rt
	srt.Table = FreeTable
	// Multiple IPv6 link-locals can't be added to the same
	// table unless the Priority differs. Different
	// LinkIndex, Src, Scope doesn't matter.
	if rt.Dst != nil && rt.Dst.IP.IsLinkLocalUnicast() {
		if debug {
			log.Printf("Forcing IPv6 priority to %v\n",
				rt.LinkIndex)
		}
		// Hack to make the kernel routes not appear identical
		srt.Priority = rt.LinkIndex
	}

	// Add for all ifindices
	MyTable := FreeTable + rt.LinkIndex

	// Add to ifindex specific table
	myrt := rt
	myrt.Table = MyTable
	// Clear any RTNH_F_LINKDOWN etc flags since add doesn't like them
	if rt.Flags != 0 {
		srt.Flags = 0
		myrt.Flags = 0
	}
	if change.Type == syscall.RTM_DELROUTE {
		if debug {
			log.Printf("Received route del %v\n", rt)
		}
		if doFreeTable {
			if err := netlink.RouteDel(&srt); err != nil {
				log.Printf("Failed to remove %v from %d: %s\n",
					srt, srt.Table, err)
			}
		}
		if err := netlink.RouteDel(&myrt); err != nil {
			log.Printf("Failed to remove %v from %d: %s\n",
				myrt, myrt.Table, err)
		}
	} else if change.Type == syscall.RTM_NEWROUTE {
		if debug {
			log.Printf("Received route add %v\n", rt)
		}
		if doFreeTable {
			if err := netlink.RouteAdd(&srt); err != nil {
				log.Printf("Failed to add %v to %d: %s\n",
					srt, srt.Table, err)
			}
		}
		if err := netlink.RouteAdd(&myrt); err != nil {
			log.Printf("Failed to add %v to %d: %s\n",
				myrt, myrt.Table, err)
		}
	}
}

// Handle an IP address change
func PbrAddrChange(change netlink.AddrUpdate) {
	changed := false
	if change.NewAddr {
		changed = IfindexToAddrsAdd(change.LinkIndex,
			change.LinkAddress)
		if changed {
			addSourceRule(change.LinkIndex, change.LinkAddress)
		}
	} else {
		changed = IfindexToAddrsDel(change.LinkIndex,
			change.LinkAddress)
		if changed {
			delSourceRule(change.LinkIndex, change.LinkAddress)
		}
	}
	if changed {
		ifname, err := IfindexToName(change.LinkIndex)
		if err != nil {
			log.Printf("PbrAddrChange IfindexToName failed for %d: %s\n",
				change.LinkIndex, err)
		} else if isUplink(ifname) {
			if debug {
				log.Printf("Address change for uplink: %v\n",
					change)
			}
			addrChangeFuncUplink(ifname)
		} else {
			if debug {
				log.Printf("Address change for non-uplink: %v\n",
					change)
			}
			if addrChangeFuncNonUplink != nil {
				addrChangeFuncNonUplink(ifname)
			}
		}
	}
}

// Handle a link being added or deleted
func PbrLinkChange(change netlink.LinkUpdate) {
	ifindex := change.Attrs().Index
	ifname := change.Attrs().Name
	switch change.Header.Type {
	case syscall.RTM_NEWLINK:
		new := IfindexToNameAdd(ifindex, ifname)
		if new {
			if isFreeUplink(ifname) {
				if debug {
					log.Printf("PbrLinkChange moving to FreeTable %s\n",
						ifname)
				}
				moveRoutesTable(0, ifindex, FreeTable)
			}
			if isUplink(ifname) {
				if debug {
					log.Printf("Link change for uplink: %s\n",
						ifname)
				}
				addrChangeFuncUplink(ifname)
			} else {
				if debug {
					log.Printf("Link change for non-uplink: %s\n",
						ifname)
				}
				if addrChangeFuncNonUplink != nil {
					addrChangeFuncNonUplink(ifname)
				}
			}

		}
	case syscall.RTM_DELLINK:
		gone := IfindexToNameDel(ifindex, ifname)
		if gone {
			if isFreeUplink(ifname) {
				flushRoutesTable(FreeTable, ifindex)
			}
			MyTable := FreeTable + ifindex
			flushRoutesTable(MyTable, 0)
			flushRules(ifindex)
			if isUplink(ifname) {
				if debug {
					log.Printf("Link change for uplink: %s\n",
						ifname)
				}
				addrChangeFuncUplink(ifname)
			} else {
				if debug {
					log.Printf("Link change for non-uplink: %s\n",
						ifname)
				}
				if addrChangeFuncNonUplink != nil {
					addrChangeFuncNonUplink(ifname)
				}
			}

		}
	}
}

var uplinkList []string     // All uplinks
var freeUplinkList []string // The subset we add to FreeTable

// Can be called to update the list.
func setFreeUplinks(freeUplinks []string) {
	if debug {
		log.Printf("setFreeUplinks(%v)\n", freeUplinks)
	}
	// Determine which ones were added; moveRoutesTable to add to free table
	for _, new := range freeUplinks {
		found := false
		for _, old := range freeUplinkList {
			if old == new {
				found = true
				break
			}
		}
		if !found {
			if ifindex, err := IfnameToIndex(new); err == nil {
				moveRoutesTable(0, ifindex, FreeTable)
			}
		}
	}
	// Determine which ones were deleted; flushRoutesTable to remove from
	// free table
	for _, old := range freeUplinkList {
		found := false
		for _, new := range freeUplinks {
			if old == new {
				found = true
				break
			}
		}
		if !found {
			if ifindex, err := IfnameToIndex(old); err == nil {
				flushRoutesTable(FreeTable, ifindex)
			}
		}
	}
	freeUplinkList = freeUplinks
}

func isFreeUplink(ifname string) bool {
	for _, fu := range freeUplinkList {
		if fu == ifname {
			return true
		}
	}
	return false
}

// Can be called to to initial set and later update the list. However,
// the caller needs to call updateListConfiglets after an update.
func setUplinks(uplinks []string) {
	uplinkList = uplinks
}

func isUplink(ifname string) bool {
	for _, u := range uplinkList {
		if u == ifname {
			return true
		}
	}
	return false
}

// ===== map from ifindex to ifname

var ifindexToName map[int]string

func IfindexToNameInit() {
	ifindexToName = make(map[int]string)
}

// Returns true if new
func IfindexToNameAdd(index int, name string) bool {
	m, ok := ifindexToName[index]
	if !ok {
		// Note that we get RTM_NEWLINK even for link changes
		// hence we don't print unless the entry is new
		log.Printf("IfindexToNameAdd index %d name %s\n", index, name)
		ifindexToName[index] = name
		// log.Printf("ifindexToName post add %v\n", ifindexToName)
		return true
	} else if m != name {
		// We get this when the vifs are created with "vif*" names
		// and then changed to "bu*" etc.
		log.Printf("IfindexToNameAdd name mismatch %s vs %s for %d\n",
			m, name, index)
		ifindexToName[index] = name
		// log.Printf("ifindexToName post add %v\n", ifindexToName)
		return false
	} else {
		return false
	}
}

// Returns true if deleted
func IfindexToNameDel(index int, name string) bool {
	m, ok := ifindexToName[index]
	if !ok {
		log.Printf("IfindexToNameDel unknown index %d\n", index)
		return false
	} else if m != name {
		log.Printf("IfindexToNameDel name mismatch %s vs %s for %d\n",
			m, name, index)
		delete(ifindexToName, index)
		// log.Printf("ifindexToName post delete %v\n", ifindexToName)
		return true
	} else {
		if debug {
			log.Printf("IfindexToNameDel index %d name %s\n",
				index, name)
		}
		delete(ifindexToName, index)
		// log.Printf("ifindexToName post delete %v\n", ifindexToName)
		return true
	}
}

func IfindexToName(index int) (string, error) {
	n, ok := ifindexToName[index]
	if ok {
		return n, nil
	}
	// Try a lookup to handle race
	link, err := netlink.LinkByIndex(index)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unknown ifindex %d", index))
	}
	name := link.Attrs().Name
	log.Printf("IfindexToName(%d) fallback lookup done: %s\n",
		index, name)
	return name, nil
}

func IfnameToIndex(ifname string) (int, error) {
	for i, name := range ifindexToName {
		if name == ifname {
			return i, nil
		}
	}
	return -1, errors.New(fmt.Sprintf("Unknown ifname %s", ifname))
}

// ===== map from ifindex to list of IP addresses

var ifindexToAddrs map[int][]net.IPNet

func IfindexToAddrsInit() {
	ifindexToAddrs = make(map[int][]net.IPNet)
}

// Returns true if added
func IfindexToAddrsAdd(index int, addr net.IPNet) bool {
	addrs, ok := ifindexToAddrs[index]
	if !ok {
		if debug {
			log.Printf("IfindexToAddrsAdd add %v for %d\n",
				addr, index)
		}
		ifindexToAddrs[index] = append(ifindexToAddrs[index], addr)
		// log.Printf("ifindexToAddrs post add %v\n", ifindexToAddrs)
		return true
	}
	found := false
	for _, a := range addrs {
		// Equal if containment in both directions?
		if a.IP.Equal(addr.IP) &&
			a.Contains(addr.IP) && addr.Contains(a.IP) {
			found = true
			break
		}
	}
	if !found {
		if debug {
			log.Printf("IfindexToAddrsAdd add %v for %d\n",
				addr, index)
		}
		ifindexToAddrs[index] = append(ifindexToAddrs[index], addr)
		// log.Printf("ifindexToAddrs post add %v\n", ifindexToAddrs)
	}
	return !found
}

// Returns true if deleted
func IfindexToAddrsDel(index int, addr net.IPNet) bool {
	addrs, ok := ifindexToAddrs[index]
	if !ok {
		log.Printf("IfindexToAddrsDel unknown index %d\n", index)
		// XXX error?
		return false
	}
	for i, a := range addrs {
		// Equal if containment in both directions?
		if a.IP.Equal(addr.IP) &&
			a.Contains(addr.IP) && addr.Contains(a.IP) {
			if debug {
				log.Printf("IfindexToAddrsDel del %v for %d\n",
					addr, index)
			}
			ifindexToAddrs[index] = append(ifindexToAddrs[index][:i],
				ifindexToAddrs[index][i+1:]...)
			// log.Printf("ifindexToAddrs post remove %v\n", ifindexToAddrs)
			// XXX should we check for zero and remove ifindex?
			return true
		}
	}
	log.Printf("IfindexToAddrsDel address not found for %d in\n",
		index, addrs)
	return false
}

func IfindexToAddrs(index int) ([]net.IPNet, error) {
	addrs, ok := ifindexToAddrs[index]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown ifindex %d", index))
	}
	return addrs, nil
}

// =====

// If ifindex is non-zero we also compare it
func flushRoutesTable(table int, ifindex int) {
	filter := netlink.Route{Table: table, LinkIndex: ifindex}
	fflags := netlink.RT_FILTER_TABLE
	if ifindex != 0 {
		fflags |= netlink.RT_FILTER_OIF
	}
	routes, err := netlink.RouteListFiltered(syscall.AF_UNSPEC,
		&filter, fflags)
	if err != nil {
		log.Fatal("RouteList failed: %v\n", err)
	}
	if debug {
		log.Printf("flushRoutesTable(%d, %d) - got %d\n",
			table, ifindex, len(routes))
	}
	for _, rt := range routes {
		if rt.Table != table {
			continue
		}
		if ifindex != 0 && rt.LinkIndex != ifindex {
			continue
		}
		if debug {
			log.Printf("flushRoutesTable(%d, %d) deleting %v\n",
				table, ifindex, rt)
		}
		if err := netlink.RouteDel(&rt); err != nil {
			log.Fatal("flushRoutesTable - RouteDel %v failed %s\n",
				rt, err)
		}
	}
}

// Used when FreeUplinks get a link added
// If ifindex is non-zero we also compare it
func moveRoutesTable(srcTable int, ifindex int, dstTable int) {
	if srcTable == 0 {
		srcTable = syscall.RT_TABLE_MAIN
	}
	filter := netlink.Route{Table: srcTable, LinkIndex: ifindex}
	fflags := netlink.RT_FILTER_TABLE
	if ifindex != 0 {
		fflags |= netlink.RT_FILTER_OIF
	}
	routes, err := netlink.RouteListFiltered(syscall.AF_UNSPEC,
		&filter, fflags)
	if err != nil {
		log.Fatal("RouteList failed: %v\n", err)
	}
	if debug {
		log.Printf("moveRoutesTable(%d, %d, %d) - got %d\n",
			srcTable, ifindex, dstTable, len(routes))
	}
	for _, rt := range routes {
		if rt.Table != srcTable {
			continue
		}
		if ifindex != 0 && rt.LinkIndex != ifindex {
			continue
		}
		art := rt
		art.Table = dstTable
		// Multiple IPv6 link-locals can't be added to the same
		// table unless the Priority differs. Different
		// LinkIndex, Src, Scope doesn't matter.
		if rt.Dst != nil && rt.Dst.IP.IsLinkLocalUnicast() {
			if debug {
				log.Printf("Forcing IPv6 priority to %v\n",
					rt.LinkIndex)
			}
			// Hack to make the kernel routes not appear identical
			art.Priority = rt.LinkIndex
		}
		// Clear any RTNH_F_LINKDOWN etc flags since add doesn't
		// like them
		if rt.Flags != 0 {
			art.Flags = 0
		}
		if debug {
			log.Printf("moveRoutesTable(%d, %d, %d) adding %v\n",
				srcTable, ifindex, dstTable, art)
		}
		if err := netlink.RouteAdd(&art); err != nil {
			log.Printf("moveRoutesTable failed to add %v to %d: %s\n",
				art, art.Table, err)
		}
	}
}

// ==== manage the ip rules

// Flush the rules we create. If ifindex is non-zero we also compare it
// Otherwise we flush the FreeTable
func flushRules(ifindex int) {
	rules, err := netlink.RuleList(syscall.AF_UNSPEC)
	if err != nil {
		log.Fatal("RuleList failed: %v\n", err)
	}
	if debug {
		log.Printf("flushRules(%d) - got %d\n", ifindex, len(rules))
	}
	for _, r := range rules {
		if ifindex == 0 && r.Table != FreeTable {
			continue
		}
		if ifindex != 0 && r.Table != FreeTable+ifindex {
			continue
		}
		if debug {
			log.Printf("flushRules: RuleDel %v\n", r)
		}
		if err := netlink.RuleDel(&r); err != nil {
			log.Fatal("flushRules - RuleDel %v failed %s\n",
				r, err)
		}
	}
}

func addSourceRule(ifindex int, p net.IPNet) {
	r := netlink.NewRule()
	r.Table = FreeTable + ifindex
	// Add rule for /32 or /128
	if p.IP.To4() != nil {
		r.Family = syscall.AF_INET
		r.Src = &net.IPNet{IP: p.IP, Mask: net.CIDRMask(32, 32)}
	} else {
		r.Family = syscall.AF_INET6
		r.Src = &net.IPNet{IP: p.IP, Mask: net.CIDRMask(128, 128)}
	}
	if debug {
		log.Printf("addSourceRule: RuleAdd %v\n", r)
	}
	// Avoid duplicate rules
	_ = netlink.RuleDel(r)
	if err := netlink.RuleAdd(r); err != nil {
		log.Printf("RuleAdd %v failed with %s\n", r, err)
		return
	}
}

func delSourceRule(ifindex int, p net.IPNet) {
	r := netlink.NewRule()
	r.Table = FreeTable + ifindex
	// Add rule for /32 or /128
	if p.IP.To4() != nil {
		r.Family = syscall.AF_INET
		r.Src = &net.IPNet{IP: p.IP, Mask: net.CIDRMask(32, 32)}
	} else {
		r.Family = syscall.AF_INET6
		r.Src = &net.IPNet{IP: p.IP, Mask: net.CIDRMask(128, 128)}
	}
	if debug {
		log.Printf("delSourceRule: RuleDel %v\n", r)
	}
	if err := netlink.RuleDel(r); err != nil {
		log.Printf("RuleDel %v failed with %s\n", r, err)
		return
	}
}
