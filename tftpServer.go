package main
/*Siddhant Vashistha---
**Simple tftp server implementation*/
import (
	"fmt"
	"net"
	"log"
	"encoding/binary"
	"os"
	"errors"
)

//Custom structures to hold packets of data for the TFTP protocol
type RRQ struct {
	opcode uint16
	filename string
	mode string
}

type WRQ struct {
	opcode uint16
	filename string
	mode string
}

type DATA struct {
	opcode uint16
	blocknum uint16
	datacont []byte
}

type ACK struct {
	opcode uint16
	blocknum uint16
}

type ERROR struct {
	opcode uint16
	errorCode uint16
	errMsg string

}

func handleUDPConnection(conn *net.UDPConn) { //Active server: handles WRQ and RRQ

	buf := make([]byte, 1024) 

	n, addr, err := conn.ReadFromUDP(buf) //Read from Client

	fmt.Println("UDP Client : ", addr)
	fmt.Println("Received from UDP Client : ", string(buf[:n]))
	currOp := binary.BigEndian.Uint16(buf[0:2]) //Get opcode from first 2 bytes
	fmt.Println("CurOp is ", currOp)

	if err != nil {
		log.Fatal(err)
	}
	
	var i int
	var fn string
	var md string
	
	for i := 2; i < len(buf); i++ {
		if buf[i]==uint8(0) {
			break
		}
		fn += string(buf[i])
	}  //Extracting filename

	i += 1
	for ; i < len(buf); i++ {
		if buf[i]==uint8(0) {
			break
		}
		md += string(buf[i])
	} //Extracting mode
	
	rrqc := RRQ{}
	wrqc := WRQ{}
	if currOp==uint16(1) { //handle off to helper function depending on type of request
		rrqc.filename = fn
		rrqc.mode = md
		rrqc.opcode = currOp 
		for {
			t:= handleRRQ(conn, rrqc.filename, currOp, addr)
			if t==true {
				break
			}
		}
	}	else if currOp==uint16(2) {
		wrqc.filename = fn
		wrqc.mode = md
		wrqc.opcode = currOp
		for {
			t := handleWRQ(conn, wrqc.filename, currOp, addr)
			if t==true {
				break
			}
		}
	}
}

func checkerread(e error, conn *net.UDPConn, addr *net.UDPAddr) (bool) {


	fmt.Println("Error is : ", e)
	fmt.Println("Server will timeout, sending error msg")
	msg := make([]byte, (2+2+len(e.Error())+1))
	err := ERROR{uint16(5), uint16(0), e.Error()}
	binary.BigEndian.PutUint16(msg, err.opcode)
	binary.BigEndian.PutUint16(msg[2:], err.errorCode) //Send to client
	errstring := e.Error()
	var x int
	for x:=0; x<len(errstring); x++ {
		msg[x+4] = errstring[x]
	}
	msg[x] = uint8(0)
	fmt.Println("Sending to client : ", msg[:])
	_, ex := conn.WriteToUDP(msg, addr)
	if ex != nil {
		log.Println(ex)
	}
	return true
}

func handleRRQ(conn *net.UDPConn, filename string, currOp uint16, addr *net.UDPAddr) (bool) { //Handles RRQs and ACKs from client
	endf := false
	blockn := uint16(0) //block number currently
	for {
		buf := make([]byte, 1024)

		n, addr, err := conn.ReadFromUDP(buf) //Get next msg from client
		fmt.Println("Received from UDP Client : ", buf[:n])
		currOp := binary.BigEndian.Uint16(buf[0:2])
		fmt.Println("CurOp is ", currOp)
		if err != nil {
			log.Fatal(err)
		}
		if currOp==uint16(1) { //if RRQ or first msg
			blockn = uint16(1)
			vals := make([]byte, 512)
			dat := DATA{uint16(3), blockn, vals}
			f, err := os.Open(filename)
			if err != nil {
				endf = checkerread(err, conn, addr)
			}
			n, err := f.Read(vals) //reading from file
			if err != nil {
				endf = checkerread(err, conn, addr)
			}
			f.Close()
			fmt.Println("read ", n, " bytes.....")
			dat.datacont = vals[:n]
			msg := make([]byte, (2+2+n))
			binary.BigEndian.PutUint16(msg, dat.opcode)
			binary.BigEndian.PutUint16(msg[2:], dat.blocknum) //Send to client
			for x:=0; x<len(vals[:n]); x++ {
				msg[x+4] = vals[x]
			}
			fmt.Println("Sending to client : ", msg[:])
			_, err = conn.WriteToUDP(msg, addr)
			if err != nil {
				log.Println(err)
			}
			//SENDING FIRST DATA SEGMENT
			if n<512 {
				endf = true
			}
		}
		if currOp==uint16(4) { //If Ack received
			bn := binary.BigEndian.Uint16(buf[2:4]) //Extract which block number Ack is received for
			if blockn==bn { //Correct execution - move on to next block
				fmt.Println("ACK RECEIVED FOR BLOCK # ", blockn)
				vals := make([]byte, 512)
				blockn = blockn+uint16(1)
				dat := DATA{uint16(3), blockn, vals}
				f, err := os.Open(filename)
				if err != nil {
					endf = checkerread(err, conn, addr)
				}
				x, eff := f.Seek((512*int64(bn)), 0) //Seek offset into file depending on block number
				fmt.Println("SEEKING AT", x)
				if eff !=nil {
					log.Println(eff)
				}
				n, err := f.Read(vals)
				if err != nil {
					endf = checkerread(err, conn, addr)
				}
				f.Close()
				fmt.Println("read ", n, " bytes.....")
				fmt.Println("from ", filename)
				dat.datacont = vals[:n]
				msg := make([]byte, (2+2+n))
				binary.BigEndian.PutUint16(msg, dat.opcode)
				binary.BigEndian.PutUint16(msg[2:], dat.blocknum)
				for x:=0; x<len(vals[:n]); x++ {
				msg[x+4] = vals[x]
				}
				fmt.Println("Sending to client : ", msg[:]) //Send to client
				fmt.Println("SENT BLOCK # ", dat.blocknum)
				_, err = conn.WriteToUDP(msg, addr)
				if err != nil {
				log.Println(err)
				}
				if n<512 { //If number of bytes read are less than 512, means that EOF has been reached. Cease transmission
					endf = true
				}
			}  else {
				endf = checkerread(errors.New("WRONG ACK RECEIVED!!!"), conn, addr)
			}
		}
		if endf==true {
			break
		}
	}
	return endf
}
func handleWRQ(conn *net.UDPConn, filename string, currOp uint16, addr *net.UDPAddr) (bool) {//function tht handles WRQs and DATA segments from the client to store to the server
	endf := false
	for {	

			buf := make([]byte, 1024)

			p, addr, err := conn.ReadFromUDP(buf)
			fmt.Println("Received from UDP Client : ", string(buf[:p])) //Get next msg from client
			currOp := binary.BigEndian.Uint16(buf[0:2])
			fmt.Println("CurOp is ", currOp)
			if err != nil {
				log.Fatal(err)
			}
			if currOp==uint16(2) { //If WRQ or first message, send ACK to begin transmission

				acknowledge := ACK{uint16(4), uint16(0)}
				
				msg := make([]byte, 4)
				binary.BigEndian.PutUint16(msg, acknowledge.opcode)
				binary.BigEndian.PutUint16(msg[2:], acknowledge.blocknum)
				fmt.Println("Sending to client : ", msg[:])
				_, err = conn.WriteToUDP(msg, addr)
				//First ACK sent, client going to send data
				if err != nil {
					log.Println(err)
				}
			} else if currOp==uint16(3) { //On receiving DATA segment from client
				bn := binary.BigEndian.Uint16(buf[2:4])
				vals := make([]byte, 512)
				dat := DATA{uint16(3), bn, vals}
				i := 0
				for i = 4; i < p; i++ {
					vals[i-4] = buf[i]
				}
				dat.datacont = vals[:]
				
				fmt.Println("Current block number is ", bn)
				pwd, err := os.Getwd()
			    if err != nil {
			        fmt.Println(err)
			        os.Exit(1)
			    }
			    fmt.Println(pwd)
				if bn==uint16(1) {//if first block, create file on server. Stores in local location
						if _, err := os.Stat(filename); err == nil {
							endf = checkerread(errors.New("FILE ALREADY EXISTS!!"), conn, addr)
							break
						} else {
							f, e := os.Create(filename)
							if e != nil {
								endf = checkerread(e, conn, addr)
							}
							f.Close()
						}					
				}

				f, err := os.OpenFile(filename, os.O_APPEND | os.O_WRONLY, 0666) //open file and append to eof
				if err != nil {
					endf = checkerread(err, conn, addr)
				}
				n, e := f.Write(vals[0:i-4])
				if e != nil {
					endf = checkerread(e, conn, addr)
				}
				fmt.Println("Writing into file: ", filename, "- ", vals[0:i-4])
				fmt.Println("Wrote ", n, " bytes")
				f.Sync()
				f.Close()

				acknowledge := ACK{uint16(4), dat.blocknum} //Reply with ACK to client to request next data chunk
				msg := make([]byte, 4)
				binary.BigEndian.PutUint16(msg, acknowledge.opcode)
				binary.BigEndian.PutUint16(msg[2:], acknowledge.blocknum)
				fmt.Println("Sending to client : ", msg[:])
				_, err = conn.WriteToUDP(msg, addr)
				//ACK TO CLIENT
				if err != nil {
					log.Println(err)
				}

				if p<516 { //If num bytes read is less than 516, that means client is done transmitting. Exit
					endf = true
					break
				}	

			}
			if endf == true {
				break
			}
	}
	return endf
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {

	hostName := "localhost"
	portNum := "1112"
	service := hostName + ":" + portNum
	//Setup server

	ServerAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		log.Fatal(err)
	}

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("UDP server up and listening on port ", portNum)
	defer ServerConn.Close()

	
	for {
			handleUDPConnection(ServerConn)
	}
}