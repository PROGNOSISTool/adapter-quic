package quictracker

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "syscall"
    "time"
)

func StartPcapCapture(conn *Connection, netInterface string) (*exec.Cmd, error) {
	bpfFilter := fmt.Sprintf("host %s and udp src or dst port %d", conn.Host.IP.String(), conn.Host.Port)
	var cmd *exec.Cmd
	if netInterface == "" {
		cmd = exec.Command("/usr/sbin/tcpdump", bpfFilter, "-w", "/tmp/pcap_quic")
	} else {
		cmd = exec.Command("/usr/sbin/tcpdump", bpfFilter, "-i", netInterface, "-w", "/tmp/pcap_quic")
	}
	err := cmd.Start()
	if err == nil {
		time.Sleep(1 * time.Second)
	}
	return cmd, err
}

func StopPcapCapture(conn *Connection, cmd *exec.Cmd) ([]byte, error) {
	time.Sleep(1 * time.Second)
	cmd.Process.Signal(syscall.SIGTERM)
	err := cmd.Wait()
	if err != nil {
		return nil, err
	}
	defer os.Remove("/tmp/pcap_quic")
	return ioutil.ReadFile("/tmp/pcap_quic")
}
