Simple TFTP server implementation
_________________________________


-To build&install, make sure you have a directory with 'tftpServer.go' inside it. Call 'go build directory/tftpServer' followed by 'go install directory/tftpServer'. This will create an executable 'tftpServer.exe' in your bin folder according to your GOPATH. Run the executable to launch the server

-You can test using 'tftp.exe' provided in the repository which is a tftp client. Usage is as follows
 'tftp.exe -i -v 127.0.0.1 1112 -m binary -c put/get file'

_________________________________

Current working conditions -
-Having limited time and ability, the server at the moment completely lacks the functionality to handle ERROR's as part of the tftp protocol, and relies on a good connection. Otherwise, results in a timeout. 
-Is able to transfer large files succesfully to the server for storage. However, it times out/gets stuck in an infinite loop while reading a large file. For example, I tried a .mp3 and .jpg
-No conditions met for retransmission either in case of bad connection
-Don't think the server can handle concurrent requests. Maybe through sockets or threading. Each instance of a request is independent of each other. 
-Files don't appear on server only after transaction. At the moment, partial files are also available :s

________________________________

Thanks for giving me this assignment. I apologize I couldn't deliver fully as per requirements, but have a fair idea of how I could've done things, with my new knowledge and a little time as well as things that I was unsure about but could have worked.

*********************************
UPDATE
-Implemented premature termination errors and fixed RRQ for large files. Able to send/receive correctly now!!
 Albeit, only from 1 client at a time. 
