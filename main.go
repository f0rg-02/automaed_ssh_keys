package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"syscall"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

/*
Written by: f0rg/Alex
Mastodon is: https://infosec.exchange/@alex_02
Buymeacoffee: https://www.buymeacoffee.com/alex_f0rg

A lot of this code I either borrowed from SO and the libraries or I
reused from other programs that I've written. The whole SO and library
codes is because I would've written the exact same thing and couldn't be
bothered to write the same code by hand. Feel free to edit it as is and reuse
the code. I don't really care too much about getting credit since we are programmers
and we copy each other's codes. I do wish that you don't explicity take full credit
as being your own and if you straight up copy and paste some of this code for lets say
your programming homework, that is on you and teachers always know when someone
is cheating and copying code.
*/

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: \n", os.Args[0])
	fmt.Println()
	flag.PrintDefaults()
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func credentials() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password + "\n"), nil
}

func ParseIps(ips string) []string {

	split_ips := strings.Split(ips, ",")
	fmt.Println("Server IPs are: ", ips)
	return split_ips
}

func ParseUser(ips []string) []string {
	var final_string []string
	for _, i := range ips {
		split_str := strings.Split(i, ":")
		final_string = append(final_string, split_str...)
	}
	return final_string
}

func ReadPub(pub_file string) string {
	data, err := os.ReadFile(pub_file)
	if err != nil {
		panic(err)
	}
	fmt.Println("Public key: " + string(data))

	return string(data)
}

func GetPasswd(client_password string) (string, string) {
	var err error
	var password string

	if client_password == "" {
		password, err = credentials()

		if err != nil {
			panic(err)
		}
		client_password = password
		return client_password, password
	} else {
		password = client_password
		return client_password, password
	}
}

func DeleteServerKeys() {
	DeleteLargeFiles("id_rsa_server")
	DeleteLargeFiles("id_rsa_server.pub")
}

func GenClientKeys() {
	MakeSSHKeyPair("id_rsa_client.pub", "id_rsa_client")
	client_public := ReadPub("id_rsa_client.pub")
	fmt.Println("Client public key:", client_public)
}

func GenServerKeys() {
	MakeSSHKeyPair("id_rsa_server.pub", "id_rsa_server")
	server_public := ReadPub("id_rsa_server.pub")
	fmt.Println("Server public key:", server_public)
}

func DownloadClientKeys(client_ip_str string, client_user_str string, client_port_str string, client_password string, password string) (string, string, string) {
	var echo_client_pub string
	// Download the file
	source_file := ".ssh/id_rsa.pub"
	destination_file := "id_rsa_client.pub"

	// Get the password for client
	fmt.Printf("Downloading public key from client for user %s and client ip %s\n", client_user_str, client_ip_str)
	fmt.Printf("Getting the password for client for user %s and client ip %s\n", client_user_str, client_ip_str)

	client_password, password = GetPasswd(client_password)

	DownloadFiles(client_ip_str+":"+client_port_str, client_user_str, client_password, source_file, destination_file)
	echo_client_pub = ReadPub("id_rsa_client.pub")

	return echo_client_pub, client_password, password
}

func main() {

	// TODO: Add an option and functionality whether to add the client public key from an existing key on the server and generate server keys.

	client_user := flag.String("client-user", "", "Specify client username.")
	client_ip := flag.String("client-ip", "", "Specify client ip address.")

	server_user := flag.String("server-user", "", "Specify server username.")
	server_ip := flag.String("server-ip", "", "Specify server ip address.")

	client_port := flag.String("client-port", "22", "Specify client SSH port.")
	server_port := flag.String("server-port", "22", "Specify server SSH port.")

	update_server := flag.Bool("update-server", false, "Specify if the server just be updated. This avoids generating new keys for the client.")
	server_ips := flag.String("server-ips", "", "Specify a list of servers with their usernames. Format is IP:Username,IP:Username,IP:Username,...")

	generate_client_keys := flag.Bool("generate-client", false, "Specify if the client keys should be generated and uploaded.")

	flag.Parse()

	if !isFlagPassed("client-user") {
		fmt.Println("Pass client username")
		usage()
		os.Exit(1)
	} else if !isFlagPassed("client-ip") {
		fmt.Println("Pass client ip")
		usage()
		os.Exit(1)
	} else if !isFlagPassed("server-user") && !isFlagPassed("update-server") {
		fmt.Println("Pass server username")
		usage()
		os.Exit(1)
	} else if !isFlagPassed("server-ip") && !isFlagPassed("update-server") {
		fmt.Println("Pass server ip")
		usage()
		os.Exit(1)
	} else if isFlagPassed("update-server") && !isFlagPassed("server-ips") {
		fmt.Println("Pass server IPs in format: IP:Username,IP:Username,IP:Username,...")
		usage()
		os.Exit(1)
	}

	var ips []string

	if isFlagPassed("server-ips") {
		ips = ParseIps(*server_ips)
		fmt.Println(ips)
	}

	// Thanks. I hate it.
	// Convert all *string to string

	// Probably a better way, but I'm tired right now.
	client_ip_str := *client_ip
	client_port_str := *client_port
	client_user_str := *client_user

	server_ip_str := *server_ip
	server_port_str := *server_port
	server_user_str := *server_user

	var password string
	var client_password string

	// Generate the keys
	if !*update_server {
		GenClientKeys()
		GenServerKeys()
		ClientSSH(client_ip_str, client_user_str, client_port_str)
		ServerSSH(server_ip_str, server_user_str, server_port_str)

		// Delete all the key files generated
		DeleteLargeFiles("id_rsa_client")
		DeleteLargeFiles("id_rsa_client.pub")
		DeleteServerKeys()
	} else {
		//var err error
		if *generate_client_keys {

			GenClientKeys()

			// Get the password for client
			fmt.Printf("Getting the password for client for user %s and client ip %s\n", client_user_str, client_ip_str)
			client_password, password = GetPasswd(client_password)

			// Transferring the keys to the client
			fmt.Println("\nTransferring over the client keys")
			UploadFiles(client_ip_str+":"+client_port_str, client_user_str, password, "id_rsa_client", ".ssh/id_rsa", "0600") // Set the private key to 600
			UploadFiles(client_ip_str+":"+client_port_str, client_user_str, password, "id_rsa_client.pub", ".ssh/id_rsa.pub", "0655")
		}

		for _, ip := range ips {
			var echo_client_pub string
			if *generate_client_keys {

				echo_client_pub = ReadPub("id_rsa_client.pub")
			} else {

				echo_client_pub, client_password, password = DownloadClientKeys(client_ip_str, client_user_str, client_port_str, client_password, password)
			}

			server_split := strings.Split(ip, ":")

			// No need for a for loop with two values.
			fmt.Println("Server: ", server_split[0])
			fmt.Println("User: ", server_split[1])

			server_ip_str = server_split[0]
			server_user_str = server_split[1]

			GenServerKeys()

			// Get the password for client
			fmt.Printf("Getting the password for client for user %s and client ip %s\n", client_user_str, client_ip_str)

			client_password, password = GetPasswd(client_password)

			// Transfer the server public key

			client, err := goph.New(client_user_str, client_ip_str, goph.Password(password))
			if err != nil {
				panic(err)
			}

			echo_server_pub := ReadPub("id_rsa_server.pub")

			fmt.Println("Tranferring the server public key to the client's authorized_keys file")

			ExecCmd(client, echo_server_pub)

			client.Close()

			// Get the password for server
			fmt.Printf("Getting the password for server for user %s and server ip %s\n", server_user_str, server_ip_str)

			password, err = credentials()

			if err != nil {
				panic(err)
			}

			// Transferring the keys to the server
			fmt.Println("\nTransferring over the server keys")
			UploadFiles(server_ip_str+":"+server_port_str, server_user_str, password, "id_rsa_server", ".ssh/id_rsa", "0600") // Set the private key to 600
			UploadFiles(server_ip_str+":"+server_port_str, server_user_str, password, "id_rsa_server.pub", ".ssh/id_rsa.pub", "0655")

			// Transfer the client public key

			client, err = goph.New(server_user_str, server_ip_str, goph.Password(password))
			if err != nil {
				panic(err)
			}

			fmt.Println("Tranferring the client public key to the server's authorized_keys file")

			ExecCmd(client, echo_client_pub)

			client.Close()
			password = client_password // So we don't end up with the server's password. Should just use a different variable.

			// Delete current server rsa keys before returning to for loop.
			DeleteServerKeys()
		}

		// Delete current client keys.
		if *generate_client_keys {
			DeleteLargeFiles("id_rsa_client")
		}
		DeleteLargeFiles("id_rsa_client.pub")
	}
	fmt.Println("Done! Double check manually by trying to connect to the ssh server as the user.")
	fmt.Println("If no password prompt is shown and you go directly to the shell, everything is working correctly.")
}

func ClientSSH(client_ip_str string, client_user_str string, client_port_str string) {
	// Get the password for client
	fmt.Printf("Getting the password for client for user %s and client ip %s\n", client_user_str, client_ip_str)

	password, err := credentials()

	if err != nil {
		panic(err)
	}

	// Transferring the keys to the client
	fmt.Println("\nTransferring over the client keys")
	UploadFiles(client_ip_str+":"+client_port_str, client_user_str, password, "id_rsa_client", ".ssh/id_rsa", "0600") // Set the private key to 600
	UploadFiles(client_ip_str+":"+client_port_str, client_user_str, password, "id_rsa_client.pub", ".ssh/id_rsa.pub", "0655")

	// Transfer the server public key

	client, err := goph.New(client_user_str, client_ip_str, goph.Password(password))
	if err != nil {
		panic(err)
	}

	echo_server_pub := ReadPub("id_rsa_server.pub")

	fmt.Println("Tranferring the server public key to the client's authorized_keys file")
	ExecCmd(client, echo_server_pub)

	client.Close()
}

func ServerSSH(server_ip_str string, server_user_str string, server_port_str string) {
	var password string
	var err error

	// Get the password for server
	fmt.Printf("Getting the password for server for user %s and server ip %s\n", server_user_str, server_ip_str)

	password, err = credentials()

	if err != nil {
		panic(err)
	}

	// Transferring the keys to the server
	fmt.Println("\nTransferring over the server keys")
	UploadFiles(server_ip_str+":"+server_port_str, server_user_str, password, "id_rsa_server", ".ssh/id_rsa", "0600") // Set the private key to 600
	UploadFiles(server_ip_str+":"+server_port_str, server_user_str, password, "id_rsa_server.pub", ".ssh/id_rsa.pub", "0655")

	// Transfer the client public key

	client, err := goph.New(server_user_str, server_ip_str, goph.Password(password))
	if err != nil {
		panic(err)
	}

	echo_client_pub := ReadPub("id_rsa_client.pub")

	fmt.Println("Tranferring the client public key to the server's authorized_keys file")
	ExecCmd(client, echo_client_pub)
	client.Close()

}

func ExecCmd(client *goph.Client, echo_pub string) {
	var out []byte
	var err error
	// Execute your command.
	// Have to do echo 'public_key' >> authorized_keys or else the file will be blank
	out, err = client.Run("echo '" + echo_pub + "' >> $HOME/.ssh/authorized_keys")

	if err != nil {
		panic(err)
	}

	// Get your output as []byte.
	fmt.Println(string(out))
}

func DeleteLargeFiles(file_src string) {

	targetFile := file_src

	// make sure we open the file with correct permission
	// otherwise we will get the
	// bad file descriptor error

	file, _ := os.OpenFile(targetFile, os.O_RDWR, 0666)

	// find out how large is the target file
	fileInfo, _ := file.Stat()

	// calculate the new slice size
	// base on how large our target file is

	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	lastPosition := 0

	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partZeroBytes := make([]byte, partSize)

		// fill out the part with zero value
		copy(partZeroBytes[:], "0")

		// over write every byte in the chunk with 0
		file.WriteAt([]byte(partZeroBytes), int64(lastPosition))

		// update last written position
		lastPosition = lastPosition + partSize
	}

	file.Close()

	// finally remove/delete our file
	os.Remove(targetFile)
}

func DownloadFiles(server string, username string, password string, source_file string, destination_file string) {
	// Use SSH key authentication from the auth package
	// we ignore the host key in this example, please change this if you use this library
	clientConfig, _ := auth.PasswordKey(username, password, ssh.InsecureIgnoreHostKey())

	// For other authentication methods see ssh.ClientConfig and ssh.AuthMethod

	// Create a new SCP client
	client := scp.NewClient(server, &clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}

	// Open a file
	f, _ := os.OpenFile(destination_file, os.O_RDWR|os.O_CREATE, 0777)

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Finaly, copy the file over
	// Usage: CopyFromFile(context, file, remotePath, permission)

	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = client.CopyFromRemote(context.Background(), f, source_file)

	if err != nil {
		fmt.Println("\nError while downloading file: ", err)
	}
}

func UploadFiles(server string, username string, password string, source_file string, destination_file string, permissions string) {
	// Use SSH key authentication from the auth package
	// we ignore the host key in this example, please change this if you use this library
	clientConfig, _ := auth.PasswordKey(username, password, ssh.InsecureIgnoreHostKey())

	// For other authentication methods see ssh.ClientConfig and ssh.AuthMethod

	// Create a new SCP client
	client := scp.NewClient(server, &clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}

	// Open a file
	f, _ := os.Open(source_file)

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Finaly, copy the file over
	// Usage: CopyFromFile(context, file, remotePath, permission)

	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = client.CopyFromFile(context.Background(), *f, destination_file, permissions)

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
}

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func MakeSSHKeyPair(public string, private string) error {
	private_key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// generate and write private key as PEM
	private_key_file, err := os.Create(private)
	if err != nil {
		return err
	}
	defer private_key_file.Close()

	private_key_pem := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(private_key)}
	if err := pem.Encode(private_key_file, private_key_pem); err != nil {
		return err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&private_key.PublicKey)
	if err != nil {
		return err
	}
	return os.WriteFile(public, ssh.MarshalAuthorizedKey(pub), 0655)
}
