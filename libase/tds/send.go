package tds

import "fmt"

// Send transmits a messages payload to the server.
//
// Return values:
// 1. Total number of bytes sent successfully (including headers)
// 2. Total number of packets sent successfully
// 3. Total number of packages sent successfully
// 4. Any occurring error
func (tds *TDSConn) Send(msg Message) (int, int, int, error) {
	totalBytes := 0
	totalPackets := 0
	totalPackages := 0

	for _, pack := range msg.Packages() {
		bytes, packets, err := tds.sendPackage(pack)
		totalBytes += bytes
		totalPackets += packets

		if err != nil {
			return totalBytes, totalPackets, totalPackages, fmt.Errorf(
				"failed to send package %d: %v", totalPackages+1, err)
		}
		totalPackages++
	}

	return totalBytes, totalPackets, totalPackages, nil
}

// sendPackage transmits a packages payload to the server.
//
// Return values:
// 1. Total number of bytes sent successfully (including headers)
// 2. Total number of packets sent successfully
// 3. Any occurring error
func (tds *TDSConn) sendPackage(pack Package) (int, int, error) {
	totalBytes := 0
	totalPackets := 0

	for packet := range pack.Packets() {
		bytes, err := tds.sendPacket(packet)
		totalBytes += bytes

		if err != nil {
			return totalBytes, totalPackets, fmt.Errorf(
				"failed to send packet %d: %v", totalPackets+1, err)
		}
		totalPackets++
	}
	return totalBytes, totalPackets, nil
}

// sendPacket transmits a packets payload to the server.
//
// Return values:
// 1. Total number of bytes sent successfully (including headers)
// 2. Any occurring error
func (tds *TDSConn) sendPacket(packet Packet) (int, error) {
	bs := packet.Bytes()

	if len(bs) != int(packet.Header.Length) {
		return 0, fmt.Errorf("packet byte length (%d) and length in header (%d) do not match",
			len(bs), packet.Header.Length)
	}

	n, err := tds.conn.Write(bs)
	if err != nil {
		return n, fmt.Errorf("failed to write packet to connection: %v", err)
	}

	return n, nil
}
