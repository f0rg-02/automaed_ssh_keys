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

func main() {

	client_user := flag.String("client-user", "", "Specify client username")
	client_ip := flag.String("client-ip", "", "Specify client ip address")

	server_user := flag.String("server-user", "", "Specify server username")
	server_ip := flag.String("server-ip", "", "Specify server ip address")

	client_port := flag.String("client-port", "22", "Specify client SSH port")
	server_port := flag.String("server-port", "22", "Specify server SSH port")

	flag.Parse()

	if !isFlagPassed("client-user") {
		usage()
		os.Exit(1)
	} else if !isFlagPassed("client-ip") {
		usage()
		os.Exit(1)
	} else if !isFlagPassed("server-user") {
		usage()
		os.Exit(1)
	} else if !isFlagPassed("server-ip") {
		usage()
		os.Exit(1)
	}

	// Thanks I hate it
	// Convert all *string to string

	// Probably a better way, but I'm tired right now.
	client_ip_str := *client_ip
	client_port_str := *client_port
	client_user_str := *client_user

	server_ip_str := *server_ip
	server_port_str := *server_port
	server_user_str := *server_user

	// Generate the keys
	MakeSSHKeyPair("id_rsa_client.pub", "id_rsa_client")
	MakeSSHKeyPair("id_rsa_server.pub", "id_rsa_server")

	server_public := ReadPub("id_rsa_server.pub")
	fmt.Println("Server public key:", server_public)

	client_public := ReadPub("id_rsa_client.pub")
	fmt.Println("Client public key:", client_public)

	// Get the password for client
	fmt.Printf("Getting the password for client for user %s and client ip %s", client_user_str, client_ip_str)

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

	// Execute your command.
	// Have to do echo 'public_key' >> authorized_keys or else the file will be blank
	out, err := client.Run("echo '" + echo_server_pub + "' >> $HOME/.ssh/authorized_keys")

	if err != nil {
		panic(err)
	}

	// Get your output as []byte.
	fmt.Println(string(out))

	if err != nil {
		panic(err)
	}

	client.Close()

	// Get the password for server
	fmt.Printf("Getting the password for server for user %s and server ip %s", server_user_str, server_ip_str)

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

	echo_client_pub := ReadPub("id_rsa_client.pub")

	fmt.Println("Tranferring the client public key to the server's authorized_keys file")

	// Execute your command.
	// Have to do echo 'public_key' >> authorized_keys or else the file will be blank
	out, err = client.Run("echo '" + echo_client_pub + "' >> $HOME/.ssh/authorized_keys")

	if err != nil {
		panic(err)
	}

	// Get your output as []byte.
	fmt.Println(string(out))

	client.Close()

	// Delete all the key files generated
	DeleteLargeFiles("id_rsa_client")
	DeleteLargeFiles("id_rsa_server")
	DeleteLargeFiles("id_rsa_client.pub")
	DeleteLargeFiles("id_rsa_server.pub")
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

func ReadPub(pub_file string) string {
	data, err := os.ReadFile(pub_file)
	if err != nil {
		panic(err)
	}
	fmt.Println("Public key: " + string(data))

	return string(data)
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
