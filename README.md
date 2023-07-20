# automaed_ssh_keys
Automate the creation of SSH keys and transfer them to the servers all from the client computer.

This tool was created to help with [automaed_ssh](https://github.com/f0rg-02/automaed_ssh) by generating the SSH keys and transfer them over to the servers so you can use Authentication via keys instead of password.

You can compile this program by running:
```
git clone https://github.com/f0rg-02/automaed_ssh_keys
cd automaed_ssh_keys && go build
```

Usage is:
```
Usage of .\auto_ssh_keys.exe: 

  -client-ip string
        Specify client ip address
  -client-port string
        Specify client SSH port (default "22")
  -client-user string
        Specify client username
  -server-ip string
        Specify server ip address
  -server-port string
        Specify server SSH port (default "22")
  -server-user string
        Specify server username
```

To run you just need to specify the ips, ports, and users. Port is optional and defaults to 22.
```
.\auto_ssh_keys.exe -client-ip <ip of client> -client-user <user ssh should connect as> -server-ip <ip of server> -server-user <user ssh should connect as>
```
I did test this on Windows against my server that I use as a lab and another Debian based distro in a VM which I was using as a "client" to run automaed_ssh. The client in this case is what automaed_ssh runs on, but it is still technically a ssh server. I just did this for a single set of servers.

#### TODO: Add an option to only add keys to server and the public key of the client to the server. Will allow to not generate new keys for the client. Might be easier to write something seperate for that?

