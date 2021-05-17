package iso

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func MakeNetwork(name string, index int) string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origin, _ := netns.Get()

	newNamespace, err := netns.NewNamed(name)

	if err != nil {
		panic(err)
	}

	netns.Set(origin)

	veth, peer := createVethPair(name)

	peerLink, err := netlink.LinkByName(peer)
	if err != nil {
		panic(err)
	}

	vethLink, err := netlink.LinkByName(veth)
	if err != nil {
		panic(err)
	}

	netlink.LinkSetNsFd(peerLink, int(newNamespace))

	vethIp := fmt.Sprintf("10.%d.1.1", index)
	peerIp := fmt.Sprintf("10.%d.1.2", index)

	addressUp(vethLink, vethIp)

	netns.Set(newNamespace)

	addressUp(peerLink, peerIp)

	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: peerLink.Attrs().Index,
		Gw:        net.ParseIP(peerIp),
	}

	err = netlink.RouteAdd(route)
	if err != nil {
		panic(err)
	}

	netns.Set(origin)

	return peerIp
}

func DestroyNetwork(name string) {
	log.Printf("Deleting network namespace %s\n", name)
	err := netns.DeleteNamed(name)
	if err != nil {
		log.Printf("No namespace for %s found", name)
	}
	vethName := fmt.Sprintf("veth%s", name)
	link, err := netlink.LinkByName(vethName)
	if err != nil {
		log.Printf("No link for %s found at %s\n", name, vethName)
	} else {
		err = netlink.LinkDel(link)
		if err != nil {
			log.Printf("Error encountered while deleting network link for %s\n", vethName)
		}
	}
}

func RunOnNamespace(namespace string, command string) (*exec.Cmd, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origin, _ := netns.Get()
	ns, err := netns.GetFromName(namespace)
	if err != nil {
		return nil, err
	}

	err = netns.Set(ns)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = netns.Set(origin)

	return cmd, err
}

func addressUp(link netlink.Link, ip string) {
	address := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ip),
			Mask: net.CIDRMask(24, 32),
		},
	}

	err := netlink.AddrAdd(link, address)
	if err != nil {
		panic(err)
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		panic(err)
	}
}

func createVethPair(name string) (string, string) {
	veth := fmt.Sprintf("veth%s", name)
	peer := fmt.Sprintf("vpeer%s", name)

	attributes := netlink.NewLinkAttrs()
	attributes.Name = veth

	pair := &netlink.Veth{
		LinkAttrs: attributes,
		PeerName:  peer,
	}

	err := netlink.LinkAdd(pair)
	if err != nil {
		panic(err)
	}

	return veth, peer
}
