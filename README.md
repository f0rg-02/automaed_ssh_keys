# automaed_ssh_keys
Automate the creation of SSH keys and transfer them to the servers all from the client computer. Please report any issues or problems you find in my golang code. I am not the best programmer.

Don't think I do anything good enough for donations, but if you would like to support me and my future work I do have a buymeacofee:

<a href="https://www.buymeacoffee.com/alex_f0rg"><img src="https://img.buymeacoffee.com/button-api/?text=Buy me a coffee&emoji=&slug=alex_f0rg&button_colour=FF5F5F&font_colour=ffffff&font_family=Cookie&outline_colour=000000&coffee_colour=FFDD00" /></a>

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
        Specify client ip address.
  -client-port string
        Specify client SSH port. (default "22")
  -client-user string
        Specify client username.
  -generate-client
        Specify if the client keys should be generated and uploaded.
  -server-ip string
        Specify server ip address.
  -server-ips string
        Specify a list of servers with their usernames. Format is IP:Username,IP:Username,IP:Username,...   
  -server-port string
        Specify server SSH port. (default "22")
  -server-user string
        Specify server username.
  -update-server
        Specify if the server just be updated. This avoids generating new keys for the client.

```

To run for single client and server you just need to specify the ips, ports, and users. Port is optional and defaults to 22.
```
.\auto_ssh_keys.exe -client-ip <ip of client> -client-user <user ssh should connect as> -server-ip <ip of server> -server-user <user ssh should connect as>
```

To run for multiple server ips you have several flag options. Required are client ip and user, and server ip(s) in format IP:Username. `-update-server` must be set to true for this to work.
```
.\auto_ssh_keys.exe -client-ip <ip of client> -client-user <username of client> -server-ips IP:Username,IP:Username,... -generate-client=true/false -update-server=true/false
```

`-generate-client` and `-update-server` defaults to false.

`-generate-client` is for if you would rather generate the client ssh keys and transfer them to the client or if you want to download prexisting public key from client to add to server's authorized keys.

I did test this on Windows against my server that I use as a lab and another Debian based distro in a VM which I was using as a "client" to run automaed_ssh. The client in this case is what automaed_ssh runs on, but it is still technically a ssh server.

#### TODO: Add the ability to add new host to known_hosts file or whatever it is called
