package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	sign               = flag.Bool("sign", false, "sign a nonce")
	gen                = flag.Bool("gen", false, "generate a new key pair")
	targetLocationDir  = flag.String("targetLocationDir", "./", "target location directory")
	nonce              = flag.String("nonce", "", "nonce to sign")
	privateKeyFilePath = flag.String("privateKeyFilePath", "./private.pem", "private key file path")
)

func init() {

}

// SignNonce signs a nonce with the private key.
func SignNonce(privateKeyFilePath, nonce string) (string, error) {
	// open private key file
	privateKeyFile, err := os.Open(privateKeyFilePath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer closeFile(privateKeyFile)

	// decode private key to []bytes
	privateKeyBytes, err := ioutil.ReadAll(privateKeyFile)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// decode private key from PEM format
	block, _ := pem.Decode(privateKeyBytes)

	// decode private key to rsa.PrivateKey
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// hashing the message with SHA256
	hashed := sha256.Sum256([]byte(nonce))

	// sign nonce with private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// encode signature to base64
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	return signatureBase64, nil
}

// closeFile is a helper function to close a file.
func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// GenerateKeys GenerateNonce generates a
// 	new Public/Private RSA key pair with 2048 bits in PEM format.
func GenerateKeys(targetLocationDir string) (string, string, error) {
	// validate targetLocationDir
	if _, err := os.Stat(targetLocationDir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("targetLocationDir does not exist")
	}

	// generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// marshal private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	// marshal public key to PEM format
	publicKey := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKey,
	})

	// save private key to file
	privateKeyFilePath := targetLocationDir + "/private.pem"
	privateKeyFile, err := os.Create(privateKeyFilePath)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	_, err = privateKeyFile.Write(privatePEM)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	defer closeFile(privateKeyFile)

	// save public key to file
	publicKeyFilePath := targetLocationDir + "/public.pem"
	publicKeyFile, err := os.Create(publicKeyFilePath)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	_, err = publicKeyFile.Write(publicPEM)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	defer closeFile(publicKeyFile)

	// return if successfully
	return string(publicPEM), string(privatePEM), nil
}

func main() {
	flag.Parse()

	// generate a new key pair
	if *gen {
		publicKey, privateKey, err := GenerateKeys(*targetLocationDir)
		if err != nil {
			return
		}
		fmt.Println("public key:")
		fmt.Println(publicKey)
		fmt.Println("private key:")
		fmt.Println(privateKey)
		return
	}

	// sign a nonce
	if *sign {
		signature, err := SignNonce(*privateKeyFilePath, *nonce)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("signature:")
		fmt.Println(signature)
		return
	}

	// print usage
	fmt.Println("Usage:")
	fmt.Println("  rsa-keygen -gen -targetLocationDir <targetLocationDir>")
	fmt.Println("  rsa-keygen -sign -nonce <nonce> -privateKeyFilePath <privateKeyFilePath>")
}
