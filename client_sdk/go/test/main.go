package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var testNetworkPath string

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func main() {

	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")

	fpcPath := os.Getenv("FPC_PATH")
	if fpcPath == "" {
		panic("FPC_PATH not set")
	}
	testNetworkPath = filepath.Join(fpcPath, "integration", "fabric-samples", "test-network")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		testNetworkPath,
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// FPC example starts here
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Get FPC Contract
	contract := fpc.GetContract(network, "ecc")

	// Setup Chaincode Enclave
	log.Println("--> Create FPC chaincode enclave: ")
	attestationParams := []string{"some params"}
	err = contract.CreateEnclave("peer0.peer1.example.com:7051", attestationParams...)
	if err != nil {
		log.Fatalf("Failed to create enclave: %v", err)
	}

	log.Println("--> QueryListEnclaveCredentials: ")
	ercc := network.GetContract("ercc")
	result, err := ercc.EvaluateTransaction("QueryListEnclaveCredentials", "ecc")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Printf("--> Result: %s\n", string(result))

	// Invoke FPC Chaincode
	log.Println("--> Invoke FPC chaincode: ")
	result, err = contract.SubmitTransaction("myFunction", "arg1", "arg2", "arg3")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Printf("--> Result: %s\n", string(result))
}
